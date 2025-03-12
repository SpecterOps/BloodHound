// Copyright 2025 Specter Ops, Inc.
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

package translate

func (s *Translator) translateMatch() error {
	currentQueryPart := s.query.CurrentPart()

	for _, part := range currentQueryPart.ConsumeCurrentPattern().Parts {
		if !part.IsTraversal {
			if err := s.translateNonTraversalPatternPart(part); err != nil {
				return err
			}
		} else {
			if err := s.translateTraversalPatternPart(part, false); err != nil {
				return err
			}
		}

		// Render this pattern part in the current query part
		if err := s.buildPatternPart(part); err != nil {
			return err
		}

		// Declare the pattern variable in scope if set
		if part.PatternBinding.Set {
			s.scope.Declare(part.PatternBinding.Value.Identifier)
		}
	}

	return s.buildPatternPredicates()
}
