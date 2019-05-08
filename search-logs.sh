aws --profile uneet-prod logs filter-log-events --log-group-name bugzilla --start-time $(date -d "-2 days" +%s000) \
	--filter-pattern '[,,,, request=*HTTP*, status_code=5**]'
