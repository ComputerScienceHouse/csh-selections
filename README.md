# Selections
[![Python 3.8](https://img.shields.io/badge/python-3.8-blue.svg)](https://www.python.org/downloads/release/python-380/)
[![Linting](https://github.com/ComputerScienceHouse/csh-selections/actions/workflows/lint.yml/badge.svg)](https://github.com/ComputerScienceHouse/csh-selections/actions/workflows/lint.yml)

Selections is a Flask / Python web app meant to help facilitate the application review process for Computer Science House.

### Running locally
There are 2 methods for configuring this project:
1. Environment variables: various variables are pulled from the environment by [config.env.py](./config.env.py). This is especially useful for running selections in a container, as configuration may be injected into the env rather than written to a file.
2. Local config file: creating a file named `config.py` will allow you to write secrets and configuration to a file. Variables specified in config.py will be passed to the application. By default this will be ignored by git. Please do not commit configuration secrets.

Configuration secrets for local development may be obtained from an RTP.

If you add configuration variables or secrets, you will need to add entries to config.env.py so they may be configured in the production environment.

 Selections requires Python 3 [(install steps)](https://docs.python-guide.org/starting/installation/) and pip. Production selections runs python 3.6.

We recommend using either the python [venv module](https://docs.python.org/3.6/library/venv.html) or [virtualenv](https://virtualenv.pypa.io/en/latest/index.html), as virtual environments help isolate your project dependencies and keep your system cleaner and more stable. Please refer to the documentation on those projects for guides on usage.


Dependiencies are specified by [requirements.txt](./requirements.txt), and may be installed through pip. Note that on some systems, pip for python 3 must be specified as `pip3`.
 ```bash
pip install -r requirements.txt
```

The application may be run locally with the following command.
```bash
flask run -h localhost -p 8080
 ```

### Profiling
To profile selections, run 
```bash
python prof.py
```

This will setup the werzeug profiler, and output profiling results in logs for every request made.

## Code standards
This project uses Pylint for linting and enforcement of code standards and styling.
Every pull request of this project will be tested by Travis CI.
Failing the linting step will block merging of your PR.
Please run pylint locally before making your PR, it will help save you some time and headache.

To run pylint use this command:
```bash
pylint selections
```

## Administration
### Preparing selections for a new year
Please see the [new year guide](./docs/new_year_guide.md).

### Deploying selections
Deployment resources and documentation are in the [deploy directory](./deploy).
