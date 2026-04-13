# API Reference

## Packages
- [landscape.config.gardener.cloud/v1alpha1](#landscapeconfiggardenercloudv1alpha1)


## landscape.config.gardener.cloud/v1alpha1




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


#### GitRepository



GitRepository contains information the Git repository containing landscape deployments and configurations.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `url` _string_ | URL specifies the Git repository URL, it can be an HTTP/S or SSH address. |  | Required: \{\} <br /> |
| `ref` _[GitRepositoryRef](#gitrepositoryref)_ | Reference specifies the Git reference to resolve and monitor for<br />changes, defaults to the 'master' branch. |  | Required: \{\} <br /> |
| `paths` _[PathConfiguration](#pathconfiguration)_ | Paths specifies the path configuration within the Git repository. |  | Required: \{\} <br /> |


#### GitRepositoryRef



GitRepositoryRef specifies the Git reference to resolve and checkout.



_Appears in:_
- [GitRepository](#gitrepository)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `branch` _string_ | Branch to check out, defaults to 'main' if no other field is defined. |  | Optional: \{\} <br /> |
| `tag` _string_ | Tag to check out, takes precedence over Branch. |  | Optional: \{\} <br /> |
| `commit` _string_ | Commit SHA to check out, takes precedence over all reference fields. |  | Optional: \{\} <br /> |




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


#### PathConfiguration



PathConfiguration contains path configuration within the Git repository.



_Appears in:_
- [GitRepository](#gitrepository)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `base` _string_ | Base is the relative path to the base directory within the Git repository. |  | Required: \{\} <br /> |
| `landscape` _string_ | Landscape is the relative path to the landscape directory within the Git repository. |  | Required: \{\} <br /> |


#### VersionConfiguration



VersionConfiguration contains configuration for versioning.



_Appears in:_
- [LandscapeKitConfiguration](#landscapekitconfiguration)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `defaultVersionsUpdateStrategy` _[DefaultVersionsUpdateStrategy](#defaultversionsupdatestrategy)_ | UpdateStrategy determines whether the versions in the default vector should be updated from the release branch on resolve.<br />Possible values are "Disabled" (default) and "ReleaseBranch". |  | Optional: \{\} <br /> |


