# Mattermost Yandex Calendar Plugin (CALDav)

[![Release](https://img.shields.io/github/v/release/LugaMuga/mattermost-yandex-calendar-plugin)](https://github.com/LugaMuga/mattermost-yandex-calendar-plugin/releases/latest)
[![HW](https://img.shields.io/github/issues/LugaMuga/mattermost-yandex-calendar-plugin/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/LugaMuga/mattermost-yandex-calendar-plugin/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

Plugin for get events from [Yandex Calendar](https://calendar.yandex.ru/) in Mattermost. Possible to improve work with any CALDav server.

## Features
- Get 10 and 1 minute notifications
- Get event updates
- Get upcoming calendar events
- Get a summary for any day you like
- Setup status 'In meeting' automatically (for server v6.2.0+)

## Installation
This plugin cannot be installed on Mattermost Cloud products, as Cloud only allows installing plugins from the marketplace.
1. Download the latest [release](https://github.com/LugaMuga/mattermost-yandex-calendar-plugin/releases) or build it from sources
2. In your Mattermost, go to **System Console** > **Plugin Management** and upload the `.tar.gz` file.
3. Add calendar bot to your team by [instruction](https://www.ibm.com/docs/en/z-chatops/1.1.0?topic=mattermost-inviting-created-bot-your-team)
4. Allow any user in Mattermost server write to anyone by DM (**System console** > **Users and teams** > **Enable users to open Direct Message channels with** > **Any user on the Mattermost server**). This is fixed in server [v6.7.0](https://github.com/mattermost/mattermost-server/pull/19713)

## Build instructions
1. Clone this repo.
2. Install [Golang](https://golang.org/doc/install), [golangci-lint](https://golangci-lint.run/usage/install/) and Automake.
3. Go into the cloned directory and run `make`. You will need to upload this to your mattermost instance through the system console and provide it a Client secret and Client ID.
4. When building is finished, the plugin file is available at `dist/com.github.lugamuga.mattermost-yandex-calendar-plugin-VERSION.tar.gz`

## Configure Yandex Calendar
Please read more [here](docs/readme.md)

## Installing For Development
Please read more [here](docker/debug/readme.md)

## Contributing
If you are interested in contributing, please fork this repo and create a pull request!

## To-Do's / Future Improvements
* Add i18n localization to plugin
* Add connect dialog
* Fix limit of 1KB in event summary in CALDav client

## Troubleshoting
* Yandex CALDav server ignore any \<comp/> and <prop-filter/> tags in calendar-query

## Authors
* **Alex Yolkin** - [Alex's Github](https://github.com/LugaMuga)

