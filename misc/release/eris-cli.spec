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
pushd %{gorepo}
GOPATH=%{_builddir} GOBIN=%{_builddir} go get -ldflags "-X github.com/eris-ltd/eris-cli/version.COMMIT=`git rev-parse --short HEAD 2>/dev/null`" github.com/eris-ltd/eris-cli/cmd/eris
popd

%install
rm -rf ${RPM_BUILD_ROOT}
mkdir -p ${RPM_BUILD_ROOT}/%{_bindir} ${RPM_BUILD_ROOT}/%{_mandir}/man1
install %{_builddir}/eris ${RPM_BUILD_ROOT}/%{_bindir}
%{_builddir}/eris man --dump > ${RPM_BUILD_ROOT}/%{_mandir}/man1/eris.1
cp %{gorepo}/README.md %{_builddir}/README
cp %{gorepo}/LICENSE.md %{_builddir}/COPYING

%files
%defattr(-, root, root, 0755)
%doc README COPYING
%{_bindir}/*
%{_mandir}/man1/*

%clean
if [ -d ${RPM_BUILD_ROOT} ]; then rm -rf $RPM_BUILD_ROOT; fi
