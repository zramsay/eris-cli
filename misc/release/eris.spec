%{define} home %{getenv:HOME}
%{define} version %{getenv:MONAX_VERSION}
%{define} release %{getenv:MONAX_RELEASE}
%{define} gorepo %{_builddir}/src/github.com/monax/cli

Summary: Monax is an application platform for building, testing, maintaining, and operating applications built to run on an ecosystem level.

Name: monax
License: GPL-3
Version: %{version}
Release: %{release}
Group: Applications/Productivity
URL: https://monax.io/docs
BuildRoot: buildroot-%{name}-%{version}-%{release}.%{_arch}

%description
Monax is an application platform for building, testing, maintaining, and operating
applications built to run on an ecosystem level. It makes it easy and simple to wrangle the dragons of smart contract blockchains.

%prep
rm -fr %{_builddir}/*
mkdir -p %{gorepo}
git clone https://github.com/monax/cli %{gorepo}

pushd %{gorepo}
git fetch origin ${MONAX_BRANCH}
git checkout ${MONAX_BRANCH}
popd

%build
pushd %{gorepo}
GOPATH=%{_builddir} GOBIN=%{_builddir} go get -ldflags "-X github.com/monax/cli/version.COMMIT=`git rev-parse --short HEAD 2>/dev/null`" github.com/monax/cli/cmd/monax
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
