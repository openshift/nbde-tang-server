FROM fedora:43
RUN dnf install -y clevis clevis-dracut clevis-systemd clevis-luks bash

ENTRYPOINT ["tail"]
CMD ["-f","/dev/null"]
