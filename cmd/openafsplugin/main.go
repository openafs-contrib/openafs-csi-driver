/**
 * Copyright 2020 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main
import (
	"flag"
	"fmt"
	"os"
	"path"
	"github.com/openafs-csi-driver/pkg/openafs"
)
func init() {
	flag.Set("logtostderr", "true")
}
var (
	endpoint          = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	driverName        = flag.String("drivername", "afscsi.openafs.org", "name of the driver")
	nodeID            = flag.String("nodeid", "", "node id")
	showVersion       = flag.Bool("version", false, "Show version.")
	version = ""
)

func main() {
	flag.Parse()
	if *showVersion {
		baseName := path.Base(os.Args[0])
		fmt.Println(baseName, version)
		return
	}
	handle()
	os.Exit(0)
}

func handle() {
	driver, err := openafs.NewOpenAFSDriver(*driverName, *nodeID, *endpoint, version)
	if err != nil {
		fmt.Printf("Failed to initialize driver: %s", err.Error())
		os.Exit(1)
	}
	driver.Run()
}
