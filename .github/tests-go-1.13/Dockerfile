FROM golang:1.13

LABEL "com.github.actions.name"="golang:1.13"
LABEL "com.github.actions.icon"="mic"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="http://github.com/gebv/ffsm"
LABEL "homepage"="http://github.com/actions"
LABEL "maintainer"="GEBV <egor.workroom@gmail.com>"

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
