import {knowledgePackages} from './knowledge-data.js';

// M1 recognizes exact single-image releases only. File-set and multi-disc
// matching are provided by the authoring tools, not guessed from files[0].
export function identifySingleImage(platform, {size, sha256}, packages=knowledgePackages) {
  const matches=[];
  for(const pkg of packages) {
    if(pkg.platform!==platform)continue;
    for(const release of pkg.releases) {
      if(release.media?.length!==1||release.fileSet)continue;
      const media=release.media[0];
      if(media.size===size&&media.sha256===sha256)matches.push({package:pkg,release});
    }
  }
  if(matches.length!==1)return {status:matches.length?'ambiguous':'unknown',labels:[]};
  const match=matches[0];
  return {status:'matched',packageId:match.package.id,revision:match.package.revision,
    sourceSHA256:match.package.sourceSHA256,releaseId:match.release.id,
    // Do not expose mutable shared generated label objects to workspace state.
    labels:match.release.labels.map(label=>({...label}))};
}

export function hasSingleImageCandidate(platform,size) {
  return knowledgePackages.some(pkg=>pkg.platform===platform&&pkg.releases.some(
    r=>!r.fileSet&&r.media?.length===1&&r.media[0].size===size));
}
