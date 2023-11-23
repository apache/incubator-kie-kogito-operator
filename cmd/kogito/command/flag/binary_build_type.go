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

package flag

// BinaryBuildType specifies what kind of binary build is being done
type BinaryBuildType string

const (
	// SourceToImageBuild specifies a s2i build (not binary build)
	SourceToImageBuild BinaryBuildType = "SourceToImageBuild"
	// BinarySpringBootJvmBuild is a Spring Boot JVM binary build
	BinarySpringBootJvmBuild BinaryBuildType = "BinarySpringBootJvmBuild"
	// BinaryQuarkusFastJarJvmBuild is a Quarkus JVM binary build with Fast Jar
	BinaryQuarkusFastJarJvmBuild BinaryBuildType = "BinaryQuarkusFastJarJvmBuild"
	// BinaryQuarkusLegacyJarJvmBuild is a Quarkus JVM binary build with Legacy Jar
	BinaryQuarkusLegacyJarJvmBuild BinaryBuildType = "BinaryQuarkusLegacyJarJvmBuild"
	// BinaryQuarkusNativeBuild is a Quarkus native binary build
	BinaryQuarkusNativeBuild BinaryBuildType = "BinaryQuarkusNativeBuild"
)
