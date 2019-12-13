# Deployment in GCP Stack

You can follow this steps to deploy and manage piia on GCP stack. It's not a definitive guide and alternative approaches can be taken to deploy in production (i.e. using `docker`), here we will show how we can deploy piia as a service.

## Generating Binary

To generate the `piia` binary to use as a service, run the following in the piia directory

```shell
env GOOS=linux buffalo build --clean-assets -v
```

We will deploy this to a Compute Engine VM running Ubuntu 18.04, thus we are generating binary for a Linux distribution.

## Database Setup

