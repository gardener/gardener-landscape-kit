# API Reference

## Packages
- [landscape.config.gardener.cloud/v1alpha1](#landscapeconfiggardenercloudv1alpha1)


## landscape.config.gardener.cloud/v1alpha1




#### BaseRepositoryConfig



BaseRepositoryConfig configures the base repository.



_Appears in:_
- [RepositoriesConfig](#repositoriesconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `target` _string_ | Target is the directory of the base content within the base repository.<br />Defaults to "./" if not specified. |  | Optional: \{\} <br /> |


#### ComponentsConfiguration



ComponentsConfiguration contains configuration for components.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `exclude` _string array_ | Exclude is a list of component names to exclude. |  | Optional: \{\} <br /> |
| `include` _string array_ | Include is a list of component names to include. |  | Optional: \{\} <br /> |


#### DefaultVersionsUpdateStrategy

_Underlying type:_ _string_

DefaultVersionsUpdateStrategy controls whether the versions in the default components vector should be updated from the release branch on generate.



_Appears in:_
- [VersionConfiguration](#versionconfiguration)

| Field | Description |
| --- | --- |
| `ReleaseBranch` | DefaultVersionsUpdateStrategyReleaseBranch indicates that the versions in the default vector should be updated from the release branch on generate.<br /> |
| `Disabled` | DefaultVersionsUpdateStrategyDisabled indicates that the versions in the default vector should not be updated on generate.<br /> |


#### GitRepositoryRef



GitRepositoryRef specifies the Git reference to resolve and checkout.



_Appears in:_
- [LandscapeRepositoryConfig](#landscaperepositoryconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `branch` _string_ | Branch to check out, defaults to 'main' if no other field is defined. |  | Optional: \{\} <br /> |
| `tag` _string_ | Tag to check out, takes precedence over Branch. |  | Optional: \{\} <br /> |
| `commit` _string_ | Commit SHA to check out, takes precedence over all reference fields. |  | Optional: \{\} <br /> |




#### LandscapeRepositoryConfig



LandscapeRepositoryConfig configures the landscape repository.



_Appears in:_
- [RepositoriesConfig](#repositoriesconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL of the landscape Git repository (http/s or ssh). |  | Required: \{\} <br /> |
| `ref` _[GitRepositoryRef](#gitrepositoryref)_ | Ref to check out (branch / tag / commit). |  | Required: \{\} <br /> |
| `baseLink` _string_ | BaseLink is the path inside the landscape repository where the base repository's content is mounted (e.g. via a Git submodule). |  | Required: \{\} <br /> |
| `target` _string_ | Target is the landscape directory within the landscape repository.<br />Defaults to "./" if not specified. |  | Optional: \{\} <br /> |


#### MergeMode

_Underlying type:_ _string_

MergeMode controls how operator overwrites are handled during three-way merge.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description |
| --- | --- |
| `Hint` | MergeModeHint annotates operator-overwritten values with a comment showing the current GLK default.<br /> |
| `Silent` | MergeModeSilent retains operator overwrites without annotation.<br /> |


#### OCMComponent



OCMComponent specifies a OCM component.



_Appears in:_
- [OCMConfig](#ocmconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ |  |  |  |
| `version` _string_ |  |  |  |


#### OCMConfig



OCMConfig contains information about root component.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `repositories` _string array_ | Repositories is a map from repository name to URL. |  |  |
| `rootComponent` _[OCMComponent](#ocmcomponent)_ | RootComponent is the configuration of the root component. |  |  |
| `originalRefs` _boolean_ | OriginalRefs is a flag to output original image references in the image vectors. |  |  |
| `ignoreMissingComponents` _boolean_ | IgnoreMissingComponents indicates whether to ignore missing components during resolution. |  | Optional: \{\} <br /> |


#### RepositoriesConfig



RepositoriesConfig describes the base and landscape repositories.
All paths inside each section are relative to that repository's root.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `base` _[BaseRepositoryConfig](#baserepositoryconfig)_ | Base configures the base repository. |  | Optional: \{\} <br /> |
| `landscape` _[LandscapeRepositoryConfig](#landscaperepositoryconfig)_ | Landscape configures the landscape repository. |  | Optional: \{\} <br /> |


#### VersionCheckMode

_Underlying type:_ _string_

VersionCheckMode controls the behavior when the tool version doesn't match the component version.



_Appears in:_
- [VersionConfiguration](#versionconfiguration)

| Field | Description |
| --- | --- |
| `Strict` | VersionCheckModeStrict indicates that version mismatches should cause an error.<br /> |
| `Warning` | VersionCheckModeWarning indicates that version mismatches should only log a warning.<br /> |


#### VersionConfiguration



VersionConfiguration contains configuration for versioning.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `defaultVersionsUpdateStrategy` _[DefaultVersionsUpdateStrategy](#defaultversionsupdatestrategy)_ | UpdateStrategy determines whether the versions in the default vector should be updated from the release branch on resolve.<br />Possible values are "Disabled" (default) and "ReleaseBranch". |  | Optional: \{\} <br /> |
| `checkMode` _[VersionCheckMode](#versioncheckmode)_ | CheckMode determines the behavior when the tool version doesn't match the gardener-landscape-kit version in the component vector.<br />Possible values are "Strict" (default) and "Warning".<br />In strict mode, version mismatches cause errors. In warning mode, only warnings are logged. |  | Optional: \{\} <br /> |


