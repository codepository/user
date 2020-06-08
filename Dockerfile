FROM scratch
ADD /user //
ADD /config.json //
EXPOSE 8080
ENTRYPOINT [ "/user" ]