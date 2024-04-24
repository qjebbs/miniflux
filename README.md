# Miniflux Fork

This is a Miniflux fork, ships with all upstream features, plus:

- New home, an article statistics page where reading starts.
- Masonry view with thumbnails.
- NSFW Feature: Show / Hide content which is `Not Safe For Work`. 
- Quickly toggle masonry / list view for every category / feed.
- Action menu, with additinal action `Mark Above as Read`.
- Save / Edit articles.
- Cache images to disk/database according to feeds settings

### Environment Variables Added

| Variable Name         | Description                                 | Default Value |
| --------------------- | ------------------------------------------- | ------------- |
| DISABLE_CACHE_SERVICE | Set the value to 1 to disable cache service | None          |
| CACHE_FREQUENCY       | Caching job frequency                       | 24 (hours)    |
| CACHE_LOCATION        | Where to save caches, "disk" or "database"  | disk          |
| DISK_STORAGE_ROOT     | The path where the disk storage located     | ./            |

> Disable HTTP service (with `DISABLE_HTTP_SERVICE`), will disable cache service on anyway.

See all other variables [here](https://miniflux.app/docs/configuration.html).

### About the NSFW Mode

The NSFW is designed to hide and skip some articles temporary in some cases, like, at work.

In the second demo below, a small dot near the `Miniflux` indicates that the NSFW mode is enabled.  

- With `NSFW` mode enabled, all feeds (and their articles) marked as NSFW will not be shown.
- Mark feeds' `NSFW` flags in the feed setting pages, before enable `NSFW` mode.
- Switch `NSFW` Mode with keyboard shorcut <kbd>Shift + N</kbd> on PC, or the `NSFW` menu on mobile.

![New home](https://user-images.githubusercontent.com/16953333/68272682-61460400-009f-11ea-9072-bd359ecfcb32.png)

![Masonry view](https://user-images.githubusercontent.com/16953333/68272214-e03a3d00-009d-11ea-9a83-5b7c4fa2c5b4.png)


Miniflux 2
==========

Miniflux is a minimalist and opinionated feed reader:

- Written in Go (Golang)
- Works only with Postgresql
- Doesn't use any ORM
- Doesn't use any complicated framework
- Use only modern vanilla Javascript (ES6 and Fetch API)
- Single binary compiled statically without dependency
- The number of features is voluntarily limited

It's simple, fast, lightweight and super easy to install.

Official website: <https://miniflux.app>

Documentation
-------------

The Miniflux documentation is available here: <https://miniflux.app/docs/> ([Man page](https://miniflux.app/miniflux.1.html))

- [Opinionated?](https://miniflux.app/opinionated.html)
- [Features](https://miniflux.app/features.html)
- [Requirements](https://miniflux.app/docs/requirements.html)
- [Installation Instructions](https://miniflux.app/docs/installation.html)
- [Upgrading to a New Version](https://miniflux.app/docs/upgrade.html)
- [Configuration](https://miniflux.app/docs/configuration.html)
- [Command Line Usage](https://miniflux.app/docs/cli.html)
- [User Interface Usage](https://miniflux.app/docs/ui.html)
- [Keyboard Shortcuts](https://miniflux.app/docs/keyboard_shortcuts.html)
- [Integration with External Services](https://miniflux.app/docs/services.html)
- [Rewrite and Scraper Rules](https://miniflux.app/docs/rules.html)
- [API Reference](https://miniflux.app/docs/api.html)
- [Development](https://miniflux.app/docs/development.html)
- [Internationalization](https://miniflux.app/docs/i18n.html)
- [Frequently Asked Questions](https://miniflux.app/faq.html)

Screenshots
-----------

Default theme:

![Default theme](https://miniflux.app/images/overview.png)

Dark theme when using keyboard navigation:

![Dark theme](https://miniflux.app/images/item-selection-black-theme.png)

Credits
-------

- Authors: Frédéric Guillot - [List of contributors](https://github.com/miniflux/v2/graphs/contributors)
- Distributed under Apache 2.0 License
