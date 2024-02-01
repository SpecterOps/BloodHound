// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/errors"

	"github.com/shirou/gopsutil/v3/mem"

	"golang.org/x/crypto/argon2"
)

const (
	SecretDigesterMethodArgon2        = "argon2"
	ErrorMalformedArgon2Digest        = errors.Error("argon2 digest is malformed")
	Argon2SaltByteLength              = 16
	Argon2DigestByteLength     uint32 = 16
	Argon2idVariant                   = "argon2id"
	mcFormatRegexPattern              = `\$([^$]+)\$v=(\d+)\$m=(\d+),t=(\d+),p=(\d+)\$([^$]+)\$(.+)`
)

// Modular Crypt Format compatible regex capture group indices
const (
	mcFormatRegexVariantCapture int = iota + 1
	mcFormatRegexVersionCapture
	mcFormatRegexMemoryCostCapture
	mcFormatRegexTimeCostCapture
	mcFormatRegexParallelismCapture
	mcFormatRegexSaltCapture
	mcFormatRegexDigestCapture
)

// Configuration constants for Argon2 tuning
const (
	maxThreadsArgon2          = 255
	numTuningSamples          = 10
	startingNumIterations     = 1
	minMemoryCostBytes        = 1 * 1024 * 1024 * 1024 // 1GB per recommendation
	minMemoryCostKibibytes    = minMemoryCostBytes / 1024
	maxMemoryCostBytes        = 10 * 1024 * 1024 * 1024 // 10GB is far above any sane recommendation so clamp it
	maxMemoryCostKibibytes    = maxMemoryCostBytes / 1024
	maxDigestersSysMemPercent = 0.2
	timeCostTuningRange       = time.Millisecond * 50
	tuningDigestContent       = "this is a test"
)

var (
	mcFormatRegex = regexp.MustCompile(mcFormatRegexPattern)
)

type ComputerSpecs struct {
	CPUThreads      uint8
	MemoryKibibytes uint32
}

// Argon2 is a frontend implementation for the recommendations specified in https://tools.ietf.org/html/draft-irtf-cfrg-argon2-04
type Argon2 struct {
	MemoryKibibytes uint32
	NumIterations   uint32
	NumThreads      uint8
}

func (s Argon2) Method() string {
	return SecretDigesterMethodArgon2
}

func (s Argon2) Digest(secret string) (SecretDigest, error) {
	hash := Argon2Digest{
		DigestVariant:   Argon2idVariant,
		Version:         argon2.Version,
		MemoryKibibytes: s.MemoryKibibytes,
		NumIterations:   s.NumIterations,
		NumThreads:      s.NumThreads,
		Salt:            make([]byte, Argon2SaltByteLength),
	}

	if _, err := rand.Read(hash.Salt); err != nil {
		return hash, err
	}

	// Actually perform the digest
	hash.Digest = argon2.IDKey([]byte(secret), hash.Salt, s.NumIterations, s.MemoryKibibytes, s.NumThreads, Argon2DigestByteLength)

	return hash, nil
}

func (s Argon2) ParseDigest(mcFormatDigest string) (SecretDigest, error) {
	var digest Argon2Digest

	if captureGroups := mcFormatRegex.FindStringSubmatch(mcFormatDigest); captureGroups == nil {
		return digest, ErrorMalformedArgon2Digest
	} else if version, err := strconv.ParseInt(captureGroups[mcFormatRegexVersionCapture], 10, 32); err != nil {
		return digest, err
	} else if memoryCost, err := strconv.ParseUint(captureGroups[mcFormatRegexMemoryCostCapture], 10, 32); err != nil {
		return digest, err
	} else if timeCost, err := strconv.ParseUint(captureGroups[mcFormatRegexTimeCostCapture], 10, 32); err != nil {
		return digest, err
	} else if parallelism, err := strconv.ParseUint(captureGroups[mcFormatRegexParallelismCapture], 10, 8); err != nil {
		return digest, err
	} else if saltBytes, err := base64.StdEncoding.DecodeString(captureGroups[mcFormatRegexSaltCapture]); err != nil {
		return digest, err
	} else if digestBytes, err := base64.StdEncoding.DecodeString(captureGroups[mcFormatRegexDigestCapture]); err != nil {
		return digest, err
	} else {
		digest.DigestVariant = captureGroups[mcFormatRegexVariantCapture]
		digest.Version = int(version)
		digest.MemoryKibibytes = uint32(memoryCost)
		digest.NumIterations = uint32(timeCost)
		digest.NumThreads = uint8(parallelism)
		digest.Salt = saltBytes
		digest.Digest = digestBytes
	}

	return digest, nil
}

type Argon2Digest struct {
	DigestVariant   string
	Version         int
	MemoryKibibytes uint32
	NumIterations   uint32
	NumThreads      uint8
	Salt            []byte
	Digest          []byte
}

func (s Argon2Digest) Validate(content string) bool {
	contentDigest := argon2.IDKey([]byte(content), s.Salt, s.NumIterations, s.MemoryKibibytes, s.NumThreads, Argon2DigestByteLength)

	for idx := 0; idx < len(contentDigest); idx++ {
		if s.Digest[idx] != contentDigest[idx] {
			return false
		}
	}

	return true
}

func (s Argon2Digest) String() string {
	output := strings.Builder{}

	// Variant
	output.WriteString("$")
	output.WriteString(s.DigestVariant)

	// Version
	output.WriteString("$v=")
	output.WriteString(strconv.Itoa(s.Version))

	// Memory cost
	output.WriteString("$m=")
	output.WriteString(strconv.FormatUint(uint64(s.MemoryKibibytes), 10))

	// Time cost
	output.WriteString(",t=")
	output.WriteString(strconv.FormatUint(uint64(s.NumIterations), 10))

	// Parallelism
	output.WriteString(",p=")
	output.WriteString(strconv.FormatUint(uint64(s.NumThreads), 10))

	// Salt
	output.WriteString("$")
	output.WriteString(base64.StdEncoding.EncodeToString(s.Salt))

	// Digest
	output.WriteString("$")
	output.WriteString(base64.StdEncoding.EncodeToString(s.Digest))

	return output.String()
}

func Tune(desiredRuntime time.Duration) (Argon2, error) {
	var (
		digester = Argon2{}
	)

	// Start with some sane defaults
	digester.NumIterations = startingNumIterations

	// Check to see what percent of the available system memory we can use is. If it is less than the recommended minimum
	// memory floor then we will start tuning at the minimum recommendation instead. If it is more than the recommended maximum
	// memory ceiling then we will start tuning at the maximum recommendation instead. Otherwise, we will start tuning at
	// the specified percentage of the available system memory.
	//
	// When assigning the memory cost value, divide by 1024. The golang argon2 interface expects the memory cost to be
	// specified in kibibytes.
	if memoryStat, err := mem.VirtualMemory(); err != nil {
		return digester, err
	} else if maxMemoryAlloc := float64(memoryStat.Total) * maxDigestersSysMemPercent; maxMemoryAlloc < minMemoryCostBytes {
		digester.MemoryKibibytes = uint32(minMemoryCostKibibytes)
	} else if maxMemoryAlloc > maxMemoryCostBytes {
		digester.MemoryKibibytes = uint32(maxMemoryCostKibibytes)
	} else {
		digester.MemoryKibibytes = uint32(maxMemoryAlloc / 1024)
	}

	// Argon2 documentation recommends using numThreads equal to double the number of available processor cores with a
	// ceiling of 255 as-per the limitations of the uint8 data type of the golang implementation.
	if numThreads := runtime.NumCPU() * 2; numThreads <= maxThreadsArgon2 {
		digester.NumThreads = uint8(numThreads)
	} else {
		digester.NumThreads = maxThreadsArgon2
	}

	return tuneWithParameters(digester, desiredRuntime)
}

func tuneWithParameters(digester Argon2, desiredRuntime time.Duration) (Argon2, error) {
	// Tell the tuning algorithm whether or not to pin the memory cost
	var (
		pinMemoryCost = false
		bestConfig    = Argon2{
			NumThreads:      digester.NumThreads,
			MemoryKibibytes: minMemoryCostKibibytes,
			NumIterations:   1,
		}
	)

	if digester.MemoryKibibytes <= minMemoryCostKibibytes {
		digester.MemoryKibibytes = minMemoryCostKibibytes
		pinMemoryCost = true
	}

	for {
		// Take 10 samples and average the time deltas
		var timeDelta time.Duration

		for sample := 0; sample < numTuningSamples; sample++ {
			startingTime := time.Now()

			if _, err := digester.Digest(tuningDigestContent); err != nil {
				return bestConfig, fmt.Errorf("failed while digesting: %w", err)
			}

			timeDelta += time.Since(startingTime)
		}

		timeDelta /= numTuningSamples

		if !pinMemoryCost {
			if timeDelta > desiredRuntime {
				// Half the memory cost if the time delta is greater than the desired runtime
				digester.MemoryKibibytes /= 2

				// Break out if we've already fallen below the floor
				if digester.MemoryKibibytes < minMemoryCostKibibytes {
					bestConfig.MemoryKibibytes = minMemoryCostKibibytes
					break
				}
			} else {
				// Pin the memory cost here and adjust upward. The desire here is to execute with the largest possible
				// memory allocation that performs within the desired specification. Once a candidate memory allocation
				// is found, further adjustments rely on increasing the number of digest iterations.
				pinMemoryCost = true
			}
		} else {
			// Copy the current tuning parameters since they perform under the desired runtime specification
			if digester.NumIterations == 1 {
				bestConfig.MemoryKibibytes = digester.MemoryKibibytes
			}

			// Attempt to increase the number of digest iterations if there's headroom, saving the last known good iteration
			if desiredRuntime-timeDelta > timeCostTuningRange {
				bestConfig.NumIterations = digester.NumIterations
				digester.NumIterations++
			} else {
				break
			}
		}
	}

	return bestConfig, nil
}
