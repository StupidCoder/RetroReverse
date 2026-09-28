#pragma once
#include "gte.h"
#include <cctype>
#include <cstring>
#include <stdexcept>
#include <string>
#include <vector>
#include <functional>
#include <array>
static u32 le32(const u8 *p) {
  return u32(p[0]) | (u32(p[1]) << 8) | (u32(p[2]) << 16) | (u32(p[3]) << 24);
}
struct Disc {
  std::vector<u8> bytes;
  size_t imageSize = 0;
  std::function<void(size_t,u8*,size_t)> reader;
  mutable std::vector<u8> cache;
  mutable size_t cacheBase = size_t(-1);
  int stride = 0, offset = 0, sectors = 0;
  u32 rootLBA = 0, rootSize = 0;
  std::string name, bootName;
  const u8 *block(int n) const {
    if (n < 0 || n >= sectors)
      throw std::runtime_error("CD sector out of range");
    if (!reader) return bytes.data() + size_t(n) * stride + offset;
    size_t base = (size_t(n) / 64) * 64 * stride;
    if (base != cacheBase) {
      cache.resize(std::min(size_t(64 * stride), imageSize - base));
      rrprof::Scope clock(3,"Disc I/O");
      reader(base, cache.data(), cache.size());
      cacheBase = base;
    }
    return cache.data() + size_t(n) * stride - base + offset;
  }
  void open(const u8 *p, size_t n) {
    if (n < 17 * 2048 || n > 0xffffffffULL)
      throw std::runtime_error("Expected a data-track BIN/ISO between 34 KiB and 4 GiB");
    imageSize = n;
    if (n % 2352 == 0 && p[0] == 0 && p[1] == 255 && p[11] == 0) {
      stride = 2352;
      if (p[15] != 1 && p[15] != 2)
        throw std::runtime_error("Unsupported raw CD sector mode");
      offset = p[15] == 1 ? 16 : 24;
    } else if (n % 2048 == 0) {
      stride = 2048;
      offset = 0;
    } else
      throw std::runtime_error("BIN/ISO sector size is invalid");
    if (!reader) bytes.assign(p, p + n);
    sectors = int(n / stride);
    auto b = block(16);
    if (b[0] != 1 || memcmp(b + 1, "CD001", 5))
      throw std::runtime_error("Missing ISO 9660 volume descriptor");
    name.assign((const char *)b + 40, 32);
    while (!name.empty() && (name.back() == ' ' || name.back() == 0))
      name.pop_back();
    rootLBA = le32(b + 158);
    rootSize = le32(b + 166);
  }
  void openFile(size_t n, std::function<void(size_t,u8*,size_t)> read) {
    reader = std::move(read);
    std::array<u8,32> head{};
    if (n < head.size()) throw std::runtime_error("Disc too short");
    reader(0, head.data(), head.size());
    open(head.data(), n);
  }
  static std::string upper(std::string s) {
    for (auto &c : s) {
      if (c == '\\')
        c = '/';
      else
        c = char(std::toupper(u8(c)));
    }
    return s;
  }
  std::vector<u8> file(std::string path) const {
    path = upper(path);
    while (!path.empty() && path[0] == '/')
      path.erase(0, 1);
    u32 lba = rootLBA, size = rootSize;
    size_t from = 0;
    while (from < path.size()) {
      auto slash = path.find('/', from);
      std::string want = path.substr(from, slash == std::string::npos ? slash : slash - from);
      bool found = false;
      if (size > imageSize || size > 32 * 1024 * 1024)
        throw std::runtime_error("Invalid ISO directory extent");
      for (u32 off = 0; off < size && !found; off += 2048) {
        auto b = block(lba + off / 2048);
        for (u32 pos = 0; pos < std::min(2048u, size - off);) {
          u32 len = b[pos];
          if (!len)
            break;
          if (len < 34 || pos + len > 2048 || 33u + b[pos + 32] > len)
            throw std::runtime_error("Malformed ISO directory record");
          const u8 *r = b + pos;
          pos += len;
          std::string id = upper(std::string((const char *)r + 33, r[32]));
          if (want.find(';') == std::string::npos)
            id = id.substr(0, id.find(';'));
          if (id != want)
            continue;
          bool isDir = r[25] & 2;
          if ((slash != std::string::npos) != isDir)
            throw std::runtime_error("ISO path type mismatch");
          lba = le32(r + 2);
          size = le32(r + 10);
          found = true;
          break;
        }
      }
      if (!found)
        throw std::runtime_error("Disc file not found: " + want);
      if (slash == std::string::npos)
        break;
      from = slash + 1;
    }
    if (size > imageSize || size > 32 * 1024 * 1024)
      throw std::runtime_error("Invalid ISO file extent");
    std::vector<u8> out(size);
    for (u32 off = 0; off < size; off += 2048) {
      auto b = block(lba + off / 2048);
      memcpy(out.data() + off, b, std::min(2048u, size - off));
    }
    return out;
  }
  std::vector<u8> boot() {
    auto cfg = file("SYSTEM.CNF");
    std::string text((char *)cfg.data(), cfg.size()), up = upper(text);
    auto at = up.find("BOOT"), eq = up.find('=', at);
    if (at == std::string::npos || eq == std::string::npos)
      throw std::runtime_error("No BOOT entry in SYSTEM.CNF");
    auto colon = text.find(':', eq);
    if (colon == std::string::npos)
      throw std::runtime_error("Invalid BOOT path");
    auto end = text.find_first_of("\r\n", colon);
    bootName = text.substr(colon + 1, end == std::string::npos ? end : end - colon - 1);
    while (!bootName.empty() && std::isspace(u8(bootName.back())))
      bootName.pop_back();
    return file(bootName);
  }
};
