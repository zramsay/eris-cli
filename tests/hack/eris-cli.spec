%{define} home %{getenv:HOME}
%{define} version %{getenv:ERIS_VERSION}
%{define} release %{getenv:ERIS_RELEASE}

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
cp %{home}/README %{_builddir}/README
cp %{home}/COPYING %{_builddir}/COPYING
cp %{home}/eris %{_builddir}/eris
mkdir --parents ${RPM_BUILD_ROOT}/%{_mandir}/man1
cp %{home}/eris.1 ${RPM_BUILD_ROOT}/%{_mandir}/man1/eris.1

%build

%install
mkdir --parents ${RPM_BUILD_ROOT}/%{_bindir}
install eris ${RPM_BUILD_ROOT}/%{_bindir}

%files
%defattr(-, root, root, 0755)
%doc README
%license COPYING
%{_mandir}/man1/*
%{_bindir}/*

%clean
if [ -d $RPM_BUILD_ROOT ]; then rm -rf $RPM_BUILD_ROOT; fi
