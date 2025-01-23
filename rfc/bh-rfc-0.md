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
