Feature: Healthcheck endpoint should inform the health of service

    Scenario: Returning a OK (200) status when health endpoint called  
        Given all downstream service is successfully communicating with the service
        When I call the health endpoint
        Then the HTTP status code should be "200"
        And the health JSON output should be:
        """
        {
            "status":"OK",
            "version":{
                "build_time":"2021-06-22T13:23:24+01:00",
                "git_commit":"8da694ccf3316a20f009f5b8b946f92b662d951f",
                "language":"go",
                "language_version":"go1.15.7",
                "version":""
            },
            "uptime":123008,
            "start_time":"2021-06-22T12:23:26.03909Z",
            "checks":[
                {
                    "name":"frontend renderer",
                    "status":"OK",
                    "status_code":200,
                    "message":"renderer is ok",
                    "last_checked":"2021-06-22T12:25:22.909399Z",
                    "last_success":"2021-06-22T12:25:22.909399Z",
                    "last_failure":"2021-06-22T12:23:55.258461Z"
                },
                {
                    "name":"Search API",
                    "status":"OK",
                    "status_code":200,
                    "message":"search-api is ok",
                    "last_checked":"2021-06-22T12:24:59.832Z",
                    "last_success":"2021-06-22T12:24:59.832Z",
                    "last_failure":"2021-06-22T12:24:28.568114Z"
                },
                {
                    "name":"API router",
                    "status":"OK",
                    "status_code":200,
                    "message":"api-router is ok",
                    "last_checked":"2021-06-22T12:25:22.909399Z",
                    "last_success":"2021-06-22T12:25:22.909399Z",
                    "last_failure":"2021-06-22T12:23:55.258461Z"
                }
            ]
        }
        """

    Scenario: Returning a WARNING (429) status when health endpoint called  
        Given one of the downstream service is struggling to communicate with the service
        When I call the health endpoint
        Then the HTTP status code should be "429"
        And the health JSON output should be:
        """
        // ADD WARNING JSON
        """

    Scenario: Returning a CRITICAL (500) status when health endpoint called  
        Given one of downstream service is unsuccessfully communicating with the service
        When I call the health endpoint
        Then the HTTP status code should be "500"
        And the health JSON output should be:
        """
        // ADD CRITICAL JSON
        """
    