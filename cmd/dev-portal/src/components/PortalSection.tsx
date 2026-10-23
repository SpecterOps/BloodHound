// Copyright 2026 Specter Ops, Inc.
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

export function PortalSection({ eyebrow, id, title }: { eyebrow: string; id: string; title: string }) {
    return (
        <section className='destination-section' id={id} aria-labelledby={`${id}-title`}>
            <span className='section-eyebrow'>{eyebrow}</span>
            <div className='section-title' id={`${id}-title`} role='heading' aria-level={2}>
                {title}
            </div>
            <p>Resources for this area of the developer portal are being organized here.</p>
            <a className='back-link' href='#top'>
                Back to portal
            </a>
        </section>
    );
}
