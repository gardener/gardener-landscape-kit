# Tests

## OCM Components Test

The package `pkg/ocm/components` includes a set of tests to ensure the proper extraction of image vectors from component descriptors.

These component descriptors are specified according to the [Open Component Model (OCM)](https://ocm.software/) specification and are stored in an OCI repository typically.
The tests validate the correct calculation of imagevector overwrites by comparing the extracted image vectors against expected results.

Imagevector overwrites are relevant in scenarios where the used OCI images and helm charts are not retrieved from the default public OCI repositories.
For example, in air-gapped environments or when using private registries, it is necessary to overwrite the default image references with those from accessible repositories.
If there are adjusted OCM component descriptors available, the image vectors can be extracted from them.

As such adjusted OCM component descriptors are private and cannot be shared publicly, the tests in this package utilize generated mock component descriptors stored locally within the testdata directory.
The generator creates a Kubernetes root OCM descriptor with Kubernetes resources for multiple versions and then uses the public OCI repository at `europe-docker.pkg.dev/gardener-project/releases` to download the OCM descriptors for gardener/gardener and one extension (shoot-cert-service extension).

### Regenerating Test Data

To regenerate these mock component descriptors, you can run `make ocm-testdata-generate` from the root of the repository.
You may want to adjust the configuration file at `pkg/ocm/components/testdata/generator/config.yaml` before running the command to specify different component versions or other parameters.
It should only be necessary if there are new features introduced in the OCM component descriptors or if some new test scenarios may not be covered by the existing test data.
It is expected that the image references checked in the tests of `pkg/ocm/components/components_test.go` have to be updated accordingly after regenerating the test data if the component versions have changed.