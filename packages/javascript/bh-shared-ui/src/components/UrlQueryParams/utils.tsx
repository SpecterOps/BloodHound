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

// Excluding top section with object information
export const createAllSectionsMap = (sectionsData: any) => {
    return sectionsData.map((sectionItem: any) => {
        const allSectionsMap: string[] = [sectionItem.label];
        // subsections for nested items
        if (sectionItem.sections) {
            const listWithSubItems = sectionItem.sections.map((nestedItem: any) => nestedItem.label);
            listWithSubItems.forEach((nestedItemString: string) => {
                allSectionsMap.push(nestedItemString);
            });
        }
        return allSectionsMap;
    });
};

// Helps in determining if its a nested label
const findLabelLocation = (allSectionsMap: string[][], label: string) => {
    const filteredArray: string[] = allSectionsMap!.find((nestedArray: string[]) =>
        nestedArray.includes(label)
    ) as string[];
    const index = filteredArray.indexOf(label);
    return { filteredArray, index };
};

const isParentOfLabel = (allSectionsMap: string[][], label: string, key: string) => {
    const { filteredArray, index } = findLabelLocation(allSectionsMap, label);
    if (index > 0) {
        return key === filteredArray[0];
    }
    return false;
};

export const collapseNonSelectedSections = (
    expandedSections: { [k: string]: boolean },
    allSectionsMap: string[][],
    label: string
) => {
    for (const [key] of Object.entries(expandedSections)) {
        const isNotParentOfSection = !isParentOfLabel(allSectionsMap, label, key);
        const isNotClickedSection = key !== label; // to not interfere with normal toggle flow
        if (isNotParentOfSection && isNotClickedSection) {
            expandedSections[key] = false;
        }
    }
};

// From string array to object
export const formatRelationshipsParams = (expandedRelationships: string[]) => {
    return expandedRelationships?.reduce((queryParamObject: { [k: string]: boolean }, relationshipsLabel: string) => {
        queryParamObject[relationshipsLabel] = true;
        return queryParamObject;
    }, {});
};

// Generates the param to add based on the label
export const manageRelationshipParams = (allSections: any, label: string) => {
    const updatedParams: string[] = [];
    const { filteredArray, index } = findLabelLocation(allSections, label);
    updatedParams.push(label);
    if (index > 0) updatedParams.unshift(filteredArray[0]);
    return updatedParams;
};
