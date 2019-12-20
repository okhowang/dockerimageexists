FROM alpine

COPY dockerimageexists .

ENTRYPOINT [ "./dockerimageexists" ]
CMD []
