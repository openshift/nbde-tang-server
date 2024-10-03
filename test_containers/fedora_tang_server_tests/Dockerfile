#
# Copyright (c) 2021 Red Hat, Inc.
# Author: Sergio Arroutbi <sarroutb@redhat.com>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

FROM fedora:42
RUN dnf install -y tang curl socat meson ninja-build git gcc-c++\
 cmake libjose-devel asciidoc http-parser-devel bash
RUN /usr/libexec/tangd-keygen /var/db/tang

COPY ["tangd-test-entrypoint", "/usr/bin/tangd-test-entrypoint"]
COPY ["tangd-test-health-check", "/usr/bin/tangd-health-check"]
RUN chmod 755 /usr/bin/tangd-test-entrypoint
RUN chmod 755 /usr/bin/tangd-health-check
RUN chmod 755 /usr/libexec/tangd
RUN ln -s /usr/libexec/tangd /usr/bin/tangd
RUN git clone https://github.com/latchset/tang /root/tang
RUN cd /root/tang && mkdir -p build && cd build && meson .. && ninja
RUN dnf remove -y git cmake http-parser-devel

ENTRYPOINT ["/usr/bin/tangd-test-entrypoint", "-d", "/root/tang/build"]
