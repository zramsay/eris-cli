%{define} home %{getenv:HOME}
%{define} version %{getenv:ERIS_VERSION}
%{define} release %{getenv:ERIS_RELEASE}
%{define} gorepo %{_builddir}/src/github.com/eris-ltd/eris-cli

Summary: Eris is a platform for building, testing, maintaining, and operating distributed applications with a blockchain backend. Eris makes it easy and simple to wrangle the dragons of smart contract blockchains.
Name: eris-cli
License: GPL-3
Version: %{version}
Release: %{release}
Group: Applications/Productivity
URL: https://docs.erisindustries.com
BuildRoot: buildroot-%{name}-%{version}-%{release}.%{_arch}

%description
Eris is a platform for building, testing,
maintaining, and operating distributed
applications with a blockchain backend.

Eris makes it easy and simple to wrangle
the dragons of smart contract blockchains.

%prep
rm -fr %{_builddir}/*
mkdir -p %{gorepo}
git clone https://github.com/eris-ltd/eris-cli %{gorepo}

pushd %{gorepo}
git fetch origin ${ERIS_BRANCH}
git checkout ${ERIS_BRANCH}
popd

%build
GOPATH=%{_builddir} GOBIN=%{_builddir} go get github.com/eris-ltd/eris-cli/cmd/eris

%install
rm -rf ${RPM_BUILD_ROOT}
mkdir -p ${RPM_BUILD_ROOT}/%{_bindir} ${RPM_BUILD_ROOT}/%{_mandir}/man1
install %{_builddir}/eris ${RPM_BUILD_ROOT}/%{_bindir}
# TODO: manual page addition is pending the issue
# https://github.com/eris-ltd/eris-cli/issues/712.
cp %{gorepo}/README.md %{_builddir}/README
cp %{gorepo}/LICENSE.md %{_builddir}/COPYING

%files
%defattr(-, root, root, 0755)
%doc README COPYING
%{_bindir}/*

%clean
if [ -d ${RPM_BUILD_ROOT} ]; then rm -rf $RPM_BUILD_ROOT; fi
