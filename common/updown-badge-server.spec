################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define debug_package  %{nil}

################################################################################

%define user_name  updown

################################################################################

Summary:        Service for generating badges for updown.io checks
Name:           updown-badge-server
Version:        1.4.0
Release:        0%{?dist}
Group:          Applications/System
License:        Apache License, Version 2.0
URL:            https://kaos.sh/updown-badge-server

Source0:        https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:      checksum.sha512

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:  golang >= 1.22

Requires:       systemd

Provides:       %{name} = %{version}-%{release}

################################################################################

%description
Service for generating badges for updown.io checks.

################################################################################

%prep
%{crc_check}

%setup -q
if [[ ! -d "%{name}/vendor" ]] ; then
  echo -e "----\nThis package requires vendored dependencies\n----"
  exit 1
elif [[ -f "%{name}/%{name}" ]] ; then
  echo -e "----\nSources must not contain precompiled binaries\n----"
  exit 1
fi

%build
pushd %{name}
  go build %{name}.go
  cp LICENSE ..
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/logrotate.d
install -dm 755 %{buildroot}%{_unitdir}
install -dm 755 %{buildroot}%{_localstatedir}/log/%{name}

install -pm 755 %{name}/%{name} \
                %{buildroot}%{_bindir}/

install -pm 644 %{name}/common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/

install -pm 644 %{name}/common/%{name}.logrotate \
                %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

install -pDm 644 %{name}/common/%{name}.service \
                 %{buildroot}%{_unitdir}/

%clean
rm -rf %{buildroot}

%pre
getent group %{user_name} >/dev/null || groupadd -r %{user_name} 2>/dev/null
getent passwd %{user_name} >/dev/null || useradd -r -M -g %{user_name} -s /sbin/nologin %{user_name} 2>/dev/null
exit 0

################################################################################

%files
%defattr(-,root,root,-)
%doc LICENSE
%attr(-,%{user_name},%{user_name}) %dir %{_localstatedir}/log/%{name}
%config(noreplace) %{_sysconfdir}/%{name}.knf
%config(noreplace) %{_sysconfdir}/logrotate.d/%{name}
%{_unitdir}/%{name}.service
%{_bindir}/%{name}

################################################################################

%changelog
* Thu Oct 10 2024 Anton Novojilov <andy@essentialkaos.com> - 1.4.0-0
- Code refactoring
- Dependencies update

* Mon Jun 24 2024 Anton Novojilov <andy@essentialkaos.com> - 1.3.2-0
- Code refactoring
- Dependencies update

* Sat Mar 30 2024 Anton Novojilov <andy@essentialkaos.com> - 1.3.1-0
- Improved support information gathering
- Code refactoring
- Dependencies update

* Fri Dec 08 2023 Anton Novojilov <andy@essentialkaos.com> - 1.3.0-0
- Change service user name from 'updownbs' to 'updown'
- Dependencies update

* Mon Sep 18 2023 Anton Novojilov <andy@essentialkaos.com> - 1.2.0-0
- Removed init script usage
- Fixed compatibility with the latest version of ek package
- Code refactoring
- Dependencies update

* Thu Mar 31 2022 Anton Novojilov <andy@essentialkaos.com> - 1.1.1-0
- Removed pkg.re usage
- Added module info
- Added Dependabot configuration

* Wed Aug 25 2021 Anton Novojilov <andy@essentialkaos.com> - 1.1.0-0
- Improved color generation for uptime and apdex badges

* Sat Aug 14 2021 Anton Novojilov <andy@essentialkaos.com> - 1.0.0-0
- Initial build for kaos-repo
