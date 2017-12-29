/** 
 *
 *
 * User: zhangbob 
 * Date: 2017/12/27 
 * Time: 下午3:52 
 */
package glog

import (
	"testing"
	"flag"
	"os"
)

func TestMain(m *testing.M) {
	flag.Set("log_dir", "/Users/zhangbob/go/src/github.com/ahlzhang/MyGlog/log/")
	flag.Parse()

	ret := m.Run()
	os.Exit(ret)
}

func TestTTT(t *testing.T) {
	//Info("===>")
	//Flush()
	//
	//Warning("===>1")
	//Flush()

	Error("===>2")

	Info("===>3")
	//Flush()
	//
	Info("===>4")
	//Flush()
}
