import {
  PromptConfig,
  RuleConfigCondition,
  RuleConfigSeverity,
  TargetCaseType,
  UserConfig,
} from "@commitlint/types";

const Config: UserConfig = {
  parserPreset: "conventional-changelog-conventionalcommits",
  rules: {
    "body-leading-blank": [RuleConfigSeverity.Disabled, "always"] as const,
    "body-max-line-length": [RuleConfigSeverity.Error, "always", 72] as const,
    "footer-empty": [RuleConfigSeverity.Error, "never"],
    "footer-leading-blank": [RuleConfigSeverity.Error, "always"] as const,
    "footer-max-line-length": [
      RuleConfigSeverity.Warning,
      "always",
      72,
    ] as const,
    "header-max-length": [RuleConfigSeverity.Error, "always", 72] as const,
    "header-trim": [RuleConfigSeverity.Error, "always"] as const,
    "subject-case": [
      RuleConfigSeverity.Disabled,
      "always",
      ["sentence-case", "start-case", "pascal-case", "upper-case"],
    ] as [RuleConfigSeverity, RuleConfigCondition, TargetCaseType[]],
    "subject-empty": [RuleConfigSeverity.Error, "never"] as const,
    "subject-full-stop": [RuleConfigSeverity.Disabled, "always", "."] as const,
    "type-case": [RuleConfigSeverity.Error, "always", "lower-case"] as const,
    "type-empty": [RuleConfigSeverity.Error, "never"] as const,
    "type-enum": [
      RuleConfigSeverity.Error,
      "always",
      ["feat", "fix", "docs", "refactor", "test", "chore", "wip"],
    ] as [RuleConfigSeverity, RuleConfigCondition, string[]],
  },
  helpUrl: "https://github.com/SpecterOps/BloodHound/blob/main/rfc/bh-rfc-2.md",
};

export default Config;

export const promptConfig: PromptConfig = {
  settings: {
    scopeEnumSeparator: "feat, fix, docs, refactor, test, chore, wip",
    enableMultipleScopes: false,
  },
  messages: {
    skip: "",
    max: "",
    min: "",
    emptyWarning: "",
    upperLimitWarning: "",
    lowerLimitWarning: "",
    ["feat"]: "",
  },
  questions: {
    type: {
      description: "Select the type of change that you're committing",
      enum: {
        feat: {
          description:
            "For introducing a new feature in the application. Denotes that the release which will include this commit will have a `MINOR` version bump if there are no other `MAJOR` version changes being applied.",
          title: "Features",
          emoji: "✨",
        },
        fix: {
          description:
            "For fixing a bug in the application. Denotes that the release which will include this commit will have a `PATCH` version bump if there are no other `MAJOR` or `MINOR` version changes being applied.",
          title: "Bug Fixes",
          emoji: "🐛",
        },
        docs: {
          description:
            "For updating existing documentation or creating new documentation.",
          title: "Documentation",
          emoji: "📚",
        },
        refactor: {
          description:
            "For changes that may change logic but does not fix a bug or introduce a new feature.",
          title: "Code Refactoring",
          emoji: "📦",
        },
        test: {
          description: "For updating existing tests or introducing new tests.",
          title: "Tests",
          emoji: "🚨",
        },
        chore: {
          description:
            "For miscellaneous changes that do not change application functionality or fit well into any of the types listed above.",
          title: "Chores",
          emoji: "♻️",
        },
        wip: {
          description:
            "A convenience type for in progress work. This is NOT an acceptable type to use for a commit that will merge into the default branch.",
          title: "Work In Progress",
          emoji: "🚧",
        },
      },
    },
    scope: {
      description:
        "What is the scope of this change (e.g. component or file name)",
    },
    subject: {
      description: "Write a short, imperative tense description of the change",
    },
    body: {
      description: "Provide a longer description of the change",
    },
    isBreaking: {
      description: "Are there any breaking changes?",
    },
    breakingBody: {
      description:
        "A BREAKING CHANGE commit requires a body. Please enter a longer description of the commit itself",
    },
    breaking: {
      description: "Describe the breaking changes",
    },
    isIssueAffected: {
      description: "Does this change affect any open issues?",
    },
    issuesBody: {
      description:
        "If issues are closed, the commit requires a body. Please enter a longer description of the commit itself",
    },
    issues: {
      description:
        'Add issue references (e.g. "fix #123", "resolves #123, Closes: 123".)',
    },
  },
};
