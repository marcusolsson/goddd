FROM scratch
ADD goddd /
ADD docs /docs
EXPOSE 8080
CMD ["/goddd"]


