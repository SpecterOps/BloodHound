gen_enforced_dependency(WorkspaceCwd, DependencyIdent, DependencyRange2, DependencyType) :-
  workspace_has_dependency(WorkspaceCwd, DependencyIdent, DependencyRange, DependencyType),
  workspace_has_dependency(OtherWorspaceCwd, DependencyIdent, DependencyRange2, DependencyType2),
  DependencyType == DependencyType2,
  DependencyRange \= DependencyRange2.
