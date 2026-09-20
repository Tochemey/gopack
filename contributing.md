Contributions are welcome!

The project adheres to [Semantic Versioning](https://semver.org)
and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).
This repo uses [Make](https://www.gnu.org/software/make/) and [Docker](https://docs.docker.com/get-docker/):
every target runs inside a dev image, so you only need those two tools installed locally.

To contribute please:

- Fork the repository
- Create a feature branch
- Submit a [pull request](https://help.github.com/articles/using-pull-requests)

### Test & Linter

Prior to submitting a [pull request](https://help.github.com/articles/using-pull-requests), please run:

```bash
make lint
make test
```

Run `make help` to list all available targets.
