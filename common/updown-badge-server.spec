################################################################################

# rpmbuilder:relative-pack true

################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define _posixroot        /
%define _root             /root
%define _bin              /bin
%define _sbin             /sbin
%define _srv              /srv
%define _home             /home
%define _opt              /opt
%define _lib32            %{_posixroot}lib
%define _lib64            %{_posixroot}lib64
%define _libdir32         %{_prefix}%{_lib32}
%define _libdir64         %{_prefix}%{_lib64}
%define _logdir           %{_localstatedir}/log
%define _rundir           %{_localstatedir}/run
%define _lockdir          %{_localstatedir}/lock/subsys
%define _cachedir         %{_localstatedir}/cache
%define _spooldir         %{_localstatedir}/spool
%define _crondir          %{_sysconfdir}/cron.d
%define _loc_prefix       %{_prefix}/local
%define _loc_exec_prefix  %{_loc_prefix}
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_libdir       %{_loc_exec_prefix}/%{_lib}
%define _loc_libdir32     %{_loc_exec_prefix}/%{_lib32}
%define _loc_libdir64     %{_loc_exec_prefix}/%{_lib64}
%define _loc_libexecdir   %{_loc_exec_prefix}/libexec
%define _loc_sbindir      %{_loc_exec_prefix}/sbin
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_datarootdir  %{_loc_prefix}/share
%define _loc_includedir   %{_loc_prefix}/include
%define _loc_mandir       %{_loc_datarootdir}/man
%define _rpmstatedir      %{_sharedstatedir}/rpm-state
%define _pkgconfigdir     %{_libdir}/pkgconfig

################################################################################

%define debug_package     %{nil}

################################################################################

%define srcdir            src/github.com/essentialkaos/%{name}

################################################################################

Summary:         Service for generating badges for updown.io checks
Name:            updown-badge-server
Version:         1.1.0
Release:         0%{?dist}
Group:           Applications/System
License:         Apache License, Version 2.0
URL:             https://github.com/essentialkaos/updown-badge-server

Source0:         https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:       checksum.sha512

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.16

Requires:        kaosv >= 2.16
Requires:        systemd

Provides:        %{name} = %{version}-%{release}

################################################################################

%description
Service for generating badges for updown.io checks.

################################################################################

%prep
%{crc_check}

%setup -q

%build
export GOPATH=$(pwd)
export GO111MODULE=auto

pushd %{srcdir}
  %{__make} %{?_smp_mflags} all
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/logrotate.d
install -dm 755 %{buildroot}%{_initddir}
install -dm 755 %{buildroot}%{_unitdir}
install -dm 755 %{buildroot}%{_logdir}/%{name}

install -pm 755 %{srcdir}/%{name} \
                %{buildroot}%{_bindir}/

install -pm 644 %{srcdir}/common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/

install -pm 755 %{srcdir}/common/%{name}.init \
                %{buildroot}%{_initddir}/%{name}

install -pm 644 %{srcdir}/common/%{name}.logrotate \
                %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

install -pDm 644 %{srcdir}/common/%{name}.service \
                 %{buildroot}%{_unitdir}/

%clean
rm -rf %{buildroot}

%pre
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || useradd -r -M -g %{name} -s /sbin/nologin %{name}
exit 0

################################################################################

%files
%defattr(-,root,root,-)
%doc LICENSE
%attr(-,%{name},%{name}) %dir %{_logdir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}.knf
%config(noreplace) %{_sysconfdir}/logrotate.d/%{name}
%{_unitdir}/%{name}.service
%{_initddir}/%{name}
%{_bindir}/%{name}

################################################################################

%changelog
* Wed Aug 25 2021 Anton Novojilov <andy@essentialkaos.com> - 1.1.0-0
- Improved color generation for uptime and apdex badges

* Sat Aug 14 2021 Anton Novojilov <andy@essentialkaos.com> - 1.0.0-0
- Initial build for kaos-repo
