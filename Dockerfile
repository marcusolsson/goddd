FROM scratch
ADD goddd /
ADD booking/docs /booking/docs
ADD tracking/docs /tracking/docs
ADD handling/docs /handling/docs
EXPOSE 8080
CMD ["/goddd"]


