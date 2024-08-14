## Testing in SocialPredict

### Overview

The SocialPredict project includes a testing framework intended to verify the functionality of the software under various market conditions. This framework is designed to handle different gaming setups and to check the correctness of the algorithms in use. Below is an explanation of how the testing process is structured and some of the practices that are followed.

### Test Setup and Configuration

A `setup.yaml` file is used to configure custom conditions and fees for the games. This file can be found [here](https://github.com/openpredictionmarkets/socialpredict/blob/main/backend/setup/setup.yaml). Users can define specific market parameters in this file, which are then tested to ensure the software behaves as expected. A mathematical overview of how these elements are used can be found [here](https://github.com/openpredictionmarkets/socialpredict/blob/main/README/MATH/README-MATH-PROB-AND-PAYOUT.md).

### Testing Guidelines

- **Pre-deployment Testing:** The goal is to ensure that all tests pass before deploying new software changes. This is particularly important when custom, non-standard conditions are introduced, as these could potentially cause issues that existing tests do not cover. If such conditions arise, it's important to document them so that administrators are aware of potential risks.

- **Infinity Avoidance:** One of the specific test conditions that have been identified is the need to avoid infinity in market calculations. An example of such a test is available [here](https://github.com/openpredictionmarkets/socialpredict/blob/main/backend/tests/market_math/marketvolume_test.go).

- **Integration with Deployment:** There is an aim to integrate testing into the deployment process, ensuring that tests are run before launching any updates. For certain tests, particularly those related to critical issues like infinity avoidance, it may be necessary to enforce them as part of the setup. Other tests that are less critical or flexible may result in a warning if they pass or fail during deployment.

### Test Environment

Currently, testing is conducted directly on the local machine rather than within a Docker container. To simplify the process for contributors, particularly those who may be less familiar with setting up local environments, there is consideration of enabling tests to be run within a Golang container. This may involve updating the backend Docker image with the necessary tools and providing documentation to guide users on how to access the container in the development environment and manually run tests.

This approach could make it easier to modify conditions and run tests without needing to fully reinitialize the SocialPredict environment.

### Contributing to Testing

Contributions to the testing framework are encouraged, whether through writing new tests, improving existing coverage, or enhancing the testing environment. The aim is to provide a testing setup that is straightforward and accessible for all contributors, regardless of their experience level.

For those interested in contributing, discussions and pull requests are welcome.
