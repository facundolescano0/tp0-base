FROM python:3.9.7-slim
COPY main.py /main.py
COPY common /common
COPY tests /tests
RUN python -m unittest tests/test_common.py
ENTRYPOINT ["/bin/sh"]
