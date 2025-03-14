---
bh-rfc: 0
title: BloodHound RFC Structure & Style Guide
authors: |
    [Lees, Dillon](dlees@specterops.io)
status: DRAFT
created: 2025-01-21
---

# BloodHound RFC Structure & Style Guide

![xkcd comic - standards](https://imgs.xkcd.com/comics/standards.png)

## 1. Overview

This document outlines how to write, structure, and maintain BloodHound RFC documents. It covers best practices for clarity, consistency, and completeness so that new RFCs can be easily drafted, reviewed, accepted, and referenced by the BloodHound team.

## 2. Motivation & Goals

RFCs should invite respectful discussion and serve as a means to encode, review, and reach decisions by consensus. As such, a set of standards is required to describe how to properly author a RFC with the following goals:

- **Consistency** - Ensure all RFCs follow a standard structure and style, making them easy to read, maintain, and discover.
- **Clarity** - Provide clear guidelines for vocabulary, phrasing, and structuring a RFC.
- **Completeness** - Ensure each RFC includes all necessary contextual information, references, and supporting materials for readers and decision makers.

## 3. When to Write a RFC

It is appropriate to author a formal RFC when:

- Standardizing processes (e.g., release procedures, development workflow, etc.).
- Introducing major features or architectural changes.
- Proposing significant changes to workflows, policies, architecture, or features.
- Documenting the impact of external systems.

Use good judgement when deciding if the overhead of a full RFC is warranted. Often, a standard Pull Request with clear commit messages will suffice for small, incremental changes.

## 4. RFC Numbering

Each RFC is assigned a positive integer value that is one greater than the greatest RFC number when the RFC is merged into the `main` branch. For example, a new RFC can expect its RFC number to be `123` if the next greatest RFC is `122` after it has been merged into the `main` branch. 

## 5. RFC Status

- **DRAFT** - The proposal is actively being developed and is open for feedback.
- **ACCEPTED** - The proposal has been accepted by appropriately assigned reviewers or has been implemented.
- **SUPERSEDED BY** <RFC identifier> - The proposal has been deprecated in favor of another RFC.

Official feedback for RFCs is received via GitHub Pull Request Comments and should be the source for committing RFC status changes.

## 6. RFC Encoding

### 6.1 Markdown

All RFCs must be authored using [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax) syntax and its [advanced formatting options](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting).

### 6.2 Storage

All RFC files must be stored in a directory at the root of the repository labeled `rfc`. Each RFC file must be named using a hyphenated combination of `bh-rfc` and its RFC numeral, and include the `md` file extension. For example, this RFC's filename is `bh-rfc-0.md` because its assigned RFC number is zero (0).

## 7. RFC Structure

BloodHound RFCs follow a consistent structure with several prescribed sections. Authors should add additional sections as needed.

### 7.1 YAML Frontmatter

The beginning of each RFC must begin with the following YAML frontmatter data:

- **bh-rfc** - The assigned RFC number according to [section 4](#4-rfc-numbering).
- **title** - A concise and descriptive name for the RFC document
- **authors** - The individuals or teams that own the proposal, formatted as a multi-line string using the [literal block scalar style](https://yaml.org/spec/1.2.2/#literal-style).
- **status** - The proposal's current status (see [section 5](#5-rfc-status)). 
- **created** - The date the proposal entered its DRAFT formatted using [RFC-3339 full-date notation](https://www.rfc-editor.org/rfc/rfc3339.html#section-5.6).

#### 7.1.1 Example

``` markdown
---
bh-rfc: 0
title: BloodHound RFC Structure & Style Guide
authors: |
    [Lees, Dillon](dlees@specterops.io)
status: DRAFT
created: 2025-01-21
---
```

### 7.2 Title

The title of the proposal should appear below the YAML frontmatter using a [first-level heading](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings). The proposal title must match the title specified in the YAML frontmatter. Optionally, one image or high level diagram may be placed immediately after the title.

#### 7.2.1 Example

``` markdown
# BloodHound RFC Structure & Style Guide

![xkcd comic - standards](https://imgs.xkcd.com/comics/standards.png)
```

### 7.3 Overview

The "overview" section follows the [title](#72-title) and must begin with the [second-level heading](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) "1. Overview". A high-level summary of what the RFC covers must follow the overview heading. The summary should be succinct and limited to a few sentences.

#### 7.3.1 Example

``` markdown
## 1. Overview

This document outlines how to write, structure, and maintain BloodHound RFC documents. It covers best practices for clarity, consistency, and completeness so that new RFCs can be easily drafted, reviewed, accepted, and referenced by the BloodHound team.
```

### 7.4 Motivation & Goals

The "motivation & goals" section follows the [overview][#73-overview] and must begin with the [second-level heading](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) "2. Motivation & Goals". Following the header, authors must include contextual information about the proposal (e.g., current state, challenges, reasons, etc.) and a bulleted list of succinct goals.

#### 7.4.1 Example

``` markdown
## 2. Motivation & Goals

RFCs should invite respectful discussion and serve as a means to encode, review, and reach decisions by consensus. As such, a set of standards is required to describe how to properly author a RFC with the following goals:

- **Consistency** - Ensure all RFCs follow a standard structure and style, making them easy to read, maintain, and discover.
- **Clarity** - Provide clear guidelines for vocabulary, phrasing, and structuring a RFC.
- **Completeness** - Ensure each RFC includes all necessary contextual information, references, and supporting materials for readers and decision makers.
```
### 7.5 Considerations

An optional "considerations" section follows the [motivation & goals section](#74-motivation--goals) and must begin with the [second-level heading](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) "3. Considerations". This section may contain several subsections titled with [third-level headings](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) with sequential sub-section numbering (e.g., 3.1, 3.2, 3.3). Each subsection should detail the specific context that was taken into account when authoring the proposal.

#### 7.5.1 Recommendations

The recommended subsections are optional and not limited to:

- **Impact on Existing Systems**
    - state how the proposal affects current workflows, systems, repositories, architecture, etc.
    - include recommendations for migration or impact mitigation
- **Security & Compliance**
    - describe security risks or enhancements
    - describe impact to compliance standards
    - describe impact to data privacy
- **Drawbacks & Alternatives**
    - outline drawbacks to the proposal
    - describe alternative solutions
    - describe any trade-offs
- **Implementation Plan**
    - describe a plan for how to introduce the proposed changes

#### 7.5.2 Example

``` markdown
## 3. Considerations

### 3.1 Impact on Existing Systems

I'm a little teapot.

### 3.2 Security & Compliance

#### 3.2.1 Risks

I am short and stout.

### 3.3 Implementation Plan

#### 3.3.1 Stages

- Here is my ladle.
- Here is my spout.
```

### 7.6 Detailed Proposal

Following either the optional [considerations section](#75-considerations) or the required [motivation & goals section](#74-motivation--goals) begins the bulk of the proposal's content. Each section must begin with a [second-level heading](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) containing the next sequential section number followed by an appropriate title. Each section may contain several subsections titled with [third-level headings](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#headings) with sequential sub-section numbering (e.g., 4.1, 4.2, 4.3). The organization of the proposal's sections and subsections is at the discretion of the primary author and reviewers.

#### 7.6.1 Example

``` markdown
## 4. Versioning

All releases conform to the [Semantic Versioning Specification](semver.org).

### 4.1 Release Candidates

Release candidates are published with the pre-release suffix `-rc` and a non-zero number corresponding to the number of attempts made to produce a stable release. For example, `v1.2.0-rc1`, `v1.2.0-rc2`, etc..

## 5. Release Schedule & Support

### 5.1 Major and Minor Releases

Major and minor releases are delivered semi-continuously on a schedule internally managed by the BloodHound Product and Engineering teams. However, end-users can anticipate a new release at the end of a given release iteration. For example, at the time of this writing a release iteration is considered three weeks; therefore, at the end of a three-week release iteration a new major or minor version will be generally available.

### 5.2 Hotfix/Patch Releases

Hotfix or patched versions are released off-schedule at the discretion of the BloodHound Product and Engineering teams. Hotfix or patched versions are published to address critical defects that can be fixed and sufficiently tested ahead of the next major or minor release.

### 5.3 Support

Due to the nature of the product's continuous delivery cycle, only the latest release is actively supported during a given release iteration. For example, if a critical defect has been identified and the BloodHound Product and Engineering teams agree to author a hotfix/patch release, the fix will exist in the patched version and subsequent versions, however, the fix will not get backported to any earlier version that has the same defect.
```

## 8. Style & Formatting Guidelines

### 8.1 Clarity & Brevity

Write in a straightforward, plain-English style, avoiding jargon unless it is well defined in the same document or appropriately referenced.

### 8.2 Headings & Subheadings

Use clear headings for each major section. All sections must be numbered for easy cross-referencing and anchor linking.

### 8.3 Diagrams & Tables

Include flowcharts, sequence diagrams, architecture diagrams, tables, etc. where helpful. Visual aids can make processes easier to understand.

### 8.4 References

#### 8.4.1 External documents

Leverage inline or reference links when referencing external documents. Do not depend on any configured [autolinks](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/configuring-autolinks-to-reference-external-resources).

##### 8.4.1.1 Example

``` markdown
This is an inline link to [this repository](https://github.com/SpecterOps/BloodHound).

This is a reference link to [this repository][1] and another link to the [SpecterOps website][2].

[1]: https://github.com/SpecterOps/BloodHound
[2]: https://specterops.io
```

#### 8.4.2 Internal references

Link directly to any section that has a heading using automatically generated anchor links. To determine the anchor link for a section use the following rules:

- Letters are converted to lower-case.
- Spaces are replaced by hyphens (`-`). All other whitespace, punctuation, or special characters are removed.
- Leading and trailing whitespace characters are removed.
- Any text formatting is removed, leaving only the contents.

More information can be found in [RFC 3986](https://www.rfc-editor.org/rfc/rfc3986#section-3.5).

## 9. Images

Images may be embedded into a RFC when appropriate. Any image that may be shared across multiple RFCs must be stored in a directory named `images`, a sibling to the RFC documents. Any image that is specific to a RFC must be stored in a directory that is a sibling to the RFC document and include [the RFC filename](#62-storage) without the `md` extension (e.g `bh-rfc-0-images`).

## 10. Maintenance & Updates

Approved RFCs may be updated directly to address typos and to make small clarifications. However, all other changes should be considered in a new RFC to supersede the existing one. Once a RFC is approved that supersedes another, the previous RFC must have its status updated to **SUPERSEDED BY <new RFC identifier>** and should remain accessible for historical reference.
