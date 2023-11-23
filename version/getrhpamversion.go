// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"flag"
	"fmt"
	"github.com/kiegroup/kogito-operator/version/rhpam"
)

var (
	prior    = flag.Bool("prior", false, "get prior product version")
	csv      = flag.Bool("csv", false, "get csv version")
	csvPrior = flag.Bool("csvPrior", false, "get prior csv version")
)

func main() {
	flag.Parse()
	if *prior {
		fmt.Println(rhpam.PriorVersion)
	} else if *csv {
		fmt.Println(rhpam.CsvVersion)
	} else if *csvPrior {
		fmt.Println(rhpam.CsvPriorVersion)
	} else {
		fmt.Println(rhpam.Version)
	}
}
