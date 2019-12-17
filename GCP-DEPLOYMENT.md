# Deployment in GCP Stack

You can follow this steps to deploy and manage piia on GCP stack. It's not a definitive guide and alternative approaches can be taken to deploy in production (i.e. using `docker`), here we will show how we can deploy piia as a service.

## Generating Binary

To generate the `piia` binary to use as a service, run the following in the piia directory

```shell
env GOOS=linux buffalo build --clean-assets -v
```

We will deploy this to a Compute Engine VM running Ubuntu 18.04, thus we are generating binary for a Linux distribution.

## GCP Project Setup

Create a new project at [GCP Console](https://console.cloud.google.com). Then,

- Go to `APIs & Services` -> `Credentials` -> `Create credentials` -> `OAuth client ID`
- Create the credential for `Web application`
- Fill up the `GOOGLE_KEY` and `GOOGLE_SECRET` in the `.env` file

```shell
GOOGLE_KEY="OAUTH_CLIENT_ID"
GOOGLE_SECRET="OAUTH_CLIENT_SECRET"
```

## Database Setup

Create a PostgreSQL database server on Cloud SQL. Make the database publicly accessible (should have public IP).

- Edit your `.env` file to add the production database credentials (temporary)

```shell
DATABASE_HOST=CLOUD_SQL_PUBLIC_IP
DATABASE_PORT=5432
DATABASE_USER=CLOUD_SQL_DB_USER
DATABASE_PASSWORD=CLOUD_SQL_DB_PASSWORD
DATABASE_NAME=CLOUD_SQL_DB_NAME
```

- From your local system, run `buffalo pop migrate`
- Run `buffalo task user:superadmin:create YOUR_EMAIL YOUR_NAME YOUR_PASSWORD` to create first super user on the site.
(i.e. `buffalo task user:superadmin:create jon@example.com "Jon Doe" "J0nD@e!23"`)
- Remove the production credentials from your `.env` file

## Storage Buckets

To export the analytics and reports, you'll need a storage bucket.

- Create a storage bucket on your choice of region in GCP (i.e. YOUR_EXPORT_BUCKET)
- Go to `IAM & admin` -> `Service accounts` -> `Create Service Account`
- Give a name, then click `Create`
- Add the `Storage object create` permission
- Generate a JSON key, then save the JSON key file to your local system. (i.e. YOUR_EXPORT_SERVICE_ACCOUNT_FILE)

> N.B. You should have a separate bucket which holds the fundoscopy images (may be on another project) as well as a service account key file with `Storage Object View` permission.

## GCE Setup

Setup the production server in GCE.

- Go to `Compute Engine` -> `VM instances`
- Click `Create`
- Give a name and select region/zone of your choice
- Select `n1-standard-1` machine type (you can select lower configuration too)
- In the `Boot Disk` section select `Change`, then select `Ubuntu 18.04 LTS`; and boot disk type to `SSD persistent disk`, click `Select`
- In the `Firewall` section, select both `Allow HTTP traffic` and `Allow HTTPS traffic`
- Click `Create`

### NGINX Setup

You can follow [this steps](https://gobuffalo.io/en/docs/deploy/proxy/#nginx) to create and start the `nginx` proxy to serve the site.

### Systemd Service Setup

Please follow [this steps](https://gobuffalo.io/en/docs/deploy/systemd/) to setup `piia` as a `systemd` service on the VM.

## Stackdriver Logging and Monitoring

- Go to Stackdriver monitoring.
- Install both `Monitoring Agent` and `Logging Agent` from https://app.google.stackdriver.com/settings/accounts/agent?project=YOUR_PROJECT_ID
- Install `Nginx plugin` from https://cloud.google.com/monitoring/agent/plugins/nginx
- Setup some `Uptime Checks` and `Alert` to your workspace. (Recommendation: at least one for website uptime check and another for instance uptime check)
