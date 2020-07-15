# How to contribute

Thank you for your interest in improving the TinyGo drivers.

We would like your help to make this project better, so we appreciate any contributions. See if one of the following descriptions matches your situation:

### New to TinyGo

We'd love to get your feedback on getting started with TinyGo. Run into any difficulty, confusion, or anything else? You are not alone. We want to know about your experience, so we can help the next people. Please open a Github issue with your questions, or you can also get in touch directly with us on our Slack channel at [https://gophers.slack.com/messages/CDJD3SUP6](https://gophers.slack.com/messages/CDJD3SUP6).

### One of the TinyGo drivers is not working as you expect

Please open a Github issue with your problem, and we will be happy to assist.

### Some specific hardware you want to use does not appear to be in the TinyGo drivers

We probably have not implemented it yet. Your contribution adding the hardware support to TinyGo would be greatly appreciated.

Please first open a Github issue. We want to help, and also make sure that there is no duplications of efforts. Sometimes what you need is already being worked on by someone else.

## How to use our Github repository

The `release` branch of this repo will always have the latest released version of the TinyGo drivers. All of the active development work for the next release will take place in the `dev` branch. The TinyGo drivers will use semantic versioning and will create a tag/release for each release.

Here is how to contribute back some code or documentation:

- Fork repo
- Create a feature branch off of the `dev` branch
- Make some useful change
- Make sure the tests still pass
- Submit a pull request against the `dev` branch.
- Be kind

## How to run tests

To run the tests:

```
make test
```
