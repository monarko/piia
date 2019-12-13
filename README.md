# PIIA (Peer)

Welcome to the source repository for PIIA. This is a brief overview of what steps you need to follow for start development.

## Requirement

- Golang (v1.13+) [We use `gomodule`, consult [https://golang.org](https://golang.org/) for installing and enabling `gomodule`]
- Buffalo Web Framework (v0.15.3+) [[http://gobuffalo.io](http://gobuffalo.io)]
- PostgreSQL Database (v9.5+)

For these, older versions might work, but can not give any guarantee.

- Node (v13.2.0)
- NPM (v6.13.1)
- Yarn (v1.15.2)

## Setup

> This version heavily uses `gomodules`, so it should be cloned in a directory outside your `$GOPATH` (if `$GOPATH` is set)

After cloning the repository, make a copy of `.env.example` as `.env`.

- Edit the `.env` file and populate your credentials.
- Run `buffalo pop migrate` to intialize database tables.
- Run `buffalo task user:superadmin:create YOUR_EMAIL YOUR_NAME YOUR_PASSWORD` to create first super user on the site. (i.e. `buffalo task user:superadmin:create jon@example.com "Jon Doe" "J0nD@e!23"`)
- Run `buffalo dev` to boot up the development server.
- Browse to [http://127.0.0.1:3000/](http://127.0.0.1:3000/) and login using your super user.

## Developments

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

## Deployment

For deployment in Google Cloud Platform, please refer to the [Deployment in GCP Stack](./GCP-DEPLOYMENT.md) document.
