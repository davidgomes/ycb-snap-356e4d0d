Fixed a data race in the AWS credentials file provider when ``watched_directory`` is configured.
Concurrent ``getCredentials()`` calls from worker threads (for example ``aws_request_signing``)
could corrupt the cached credential strings and crash Envoy. Credential refresh timing is
unchanged.
