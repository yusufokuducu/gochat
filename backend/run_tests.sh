#!/bin/bash
set -e

# Run pytest with coverage report
pytest --cov=app --cov-report=term-missing tests/
