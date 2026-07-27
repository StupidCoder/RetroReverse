package main
import ("fmt";"os";"retroreverse.com/tools/platform/gc")
func main(){
 d,err:=gc.Open(os.Args[1]); if err!=nil{panic(err)}
 defer d.Close()
 for _,f:=range d.FST.Entries{ if !f.Dir { fmt.Printf("%9d  %s\n",f.Size,f.Path) } }
}
