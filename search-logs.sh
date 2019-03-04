aws --profile uneet-prod logs filter-log-events --log-group-name bugzilla --start-time $(date -d "-1 hour" +%s000) --filter-pattern '[..., request = *HTTP*, status_code = 5**, , ,]'

#aws --profile uneet-prod logs get-log-events --log-group-name bugzilla --log-stream-name 977fb15184c19468609dfe67cf217eb92894437705edc383870b3c10e58e9828
