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

export const downloadFile = ({ data, fileName, fileType }: { data: any; fileName: string; fileType: string }) => {
    const blob = new Blob([data], { type: fileType });
    // create an anchor tag and dispatch a click event on it to trigger download
    const a = document.createElement('a');
    a.download = fileName;
    a.href = window.URL.createObjectURL(blob);
    const clickEvent = new MouseEvent('click', {
        view: window,
        bubbles: true,
        cancelable: true,
    });
    a.dispatchEvent(clickEvent);
    a.remove();
};

export const exportToJson = (e: React.MouseEvent<Element, MouseEvent>, data: any) => {
    e.preventDefault();
    downloadFile({
        data: JSON.stringify(data),
        fileName: 'bh-graph.json',
        fileType: 'text/json',
    });
};
