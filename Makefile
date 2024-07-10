arch=amd64
tarname=mattermost-${version}-linux-${arch}
install_path=/opt/mattermost
use_docker=true

user_id=$(shell id -u)
group_id=$(shell id -g)

all: mattermost mattermost-omnibus


all-nightly: mattermost-omnibus-nightly


include mmomni/build.mk
include omnitests/build.mk


check:
ifndef version
	$(error version is not set. Please append version=X.Y.Z to the make command)
endif

ifndef revision
	$(error revision is not set. Please append revision=X to the make command)
endif


$(tarname).tar.gz:
	@echo "Downloading Mattermost version v${version}"
	wget https://releases.mattermost.com/${version}/${tarname}.tar.gz


download: check $(tarname).tar.gz


clean: mmomni-clean
	@echo "Cleaning generated files"
	@rm -f *.tar.gz
	@rm -f *.deb
	@rm -rf build


mattermost_$(version)-$(revision).deb: build_path=build/mattermost/mattermost_${version}-${revision}
mattermost_$(version)-$(revision).deb:
	@echo "Unpacking Mattermost v${version}"
	rm -rf ${build_path}
	mkdir -p ${build_path}/${install_path}
	tar xf ${tarname}.tar.gz --strip-components=1 -C ${build_path}/${install_path}

	@echo "Renaming config file"
	mv ${build_path}/opt/mattermost/config/config.json \
	   ${build_path}/opt/mattermost/config/config.defaults.json

	@echo "Creating binary links"
	mkdir -p ${build_path}/usr/bin
	ln -s ${install_path}/bin/mattermost ${build_path}/usr/bin/
	ln -s ${install_path}/bin/platform ${build_path}/usr/bin/
	ln -s ${install_path}/bin/mmctl ${build_path}/usr/bin/

	@echo "Copying systemd service"
	mkdir -p ${build_path}/lib/systemd/system
	cp mattermost/files/mattermost.service ${build_path}/lib/systemd/system/

	@echo "Copying debian files"
	cp -r mattermost/DEBIAN ${build_path}/
	sed -i -e "s/%%version%%/${version}-${revision}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/%%arch%%/${arch}/g" ${build_path}/DEBIAN/control

	@echo "Building mattermost package"
ifeq ($(use_docker), true)
	docker run -it -v ${PWD}/build/mattermost:/builder/package -w /builder/package -u ${user_id}:${group_id} debian:latest dpkg-deb --build mattermost_${version}-${revision}
else
	cd ${PWD}/build/mattermost; \
	dpkg-deb --build mattermost_${version}-${revision};
endif
	cp build/mattermost/mattermost_${version}-${revision}.deb .

	@echo "Build for Mattermost v${version}-${revision} succeeded"


mattermost: download mattermost_$(version)-$(revision).deb


mattermost-omnibus_$(version)-$(revision)_%.deb: release=$*
mattermost-omnibus_$(version)-$(revision)_%.deb: build_path=build/mattermost-omnibus/${release}/mattermost-omnibus_${version}-${revision}
mattermost-omnibus_$(version)-$(revision)_%.deb: dependencies=$(shell cat "mattermost-omnibus/files/${release}_dependencies")
mattermost-omnibus_$(version)-$(revision)_%.deb:
	@echo "Creating base directory for Omnibus v${version}-${revision} ${release}"
	rm -rf ${build_path}
	mkdir -p ${build_path}/${install_path}/mmomni/bin
	mkdir -p ${build_path}/etc/systemd/system/mattermost.service.d/

	@echo "Copying mmomni CLI"
	cp bin/mmomni ${build_path}/${install_path}/mmomni/bin/
	cp -r mmomni/ansible ${build_path}/${install_path}/mmomni/ansible

	@echo "Creating binary links"
	mkdir -p ${build_path}/usr/local/bin
	ln -s ${install_path}/mmomni/bin/mmomni ${build_path}/usr/local/bin/

	@echo "Copying debian files"
	cp -r mattermost-omnibus/DEBIAN ${build_path}/
	sed -i -e "s/%%dependencies%%/${dependencies}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/%%version%%/${version}-${revision}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/%%arch%%/${arch}/g" ${build_path}/DEBIAN/control

	@echo "Copy mattermost service overrides for omnibus"
	cp mattermost/files/omnibus-overrides.conf ${build_path}/etc/systemd/system/mattermost.service.d/

	@echo "Building Mattermost Omnibus v${version}-${revision} ${release} package"
ifeq ($(use_docker), true)
	docker run -it -v ${PWD}/build/mattermost-omnibus/${release}:/builder/package -w /builder/package -u ${user_id}:${group_id} debian:latest dpkg-deb --build mattermost-omnibus_${version}-${revision}
else
	cd ${PWD}/build/mattermost-omnibus/${release}; \
	dpkg-deb --build mattermost-omnibus_${version}-${revision};
endif
	cp build/mattermost-omnibus/${release}/mattermost-omnibus_${version}-${revision}.deb mattermost-omnibus_${version}-${revision}_${release}.deb

	@echo "Build for Mattermost Omnibus v${version}-${revision} ${release} succeeded"


mattermost-omnibus-nightly_$(version)-$(revision)_%.deb: release=$*
mattermost-omnibus-nightly_$(version)-$(revision)_%.deb: build_path=build/mattermost-omnibus/${release}/mattermost-omnibus-nightly_${version}-${revision}
mattermost-omnibus-nightly_$(version)-$(revision)_%.deb: dependencies=$(shell cat "mattermost-omnibus/files/${release}_dependencies")
mattermost-omnibus-nightly_$(version)-$(revision)_%.deb:
	@echo "Creating base directory for Omnibus v${version}-${revision} ${release}"
	rm -rf ${build_path}
	mkdir -p ${build_path}/${install_path}/mmomni/bin
	mkdir -p ${build_path}/etc/systemd/system/mattermost.service.d/

	@echo "Copying mmomni CLI"
	cp bin/mmomni ${build_path}/${install_path}/mmomni/bin/
	cp -r mmomni/ansible ${build_path}/${install_path}/mmomni/ansible

	@echo "Creating binary links"
	mkdir -p ${build_path}/usr/local/bin
	ln -s ${install_path}/mmomni/bin/mmomni ${build_path}/usr/local/bin/

	@echo "Copying debian files"
	cp -r mattermost-omnibus/DEBIAN ${build_path}/
	sed -i -e "s/%%dependencies%%/${dependencies}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/%%version%%/${version}-${revision}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/%%arch%%/${arch}/g" ${build_path}/DEBIAN/control
	sed -i -e "s/Package: mattermost-omnibus/Package: mattermost-omnibus-nightly/g" ${build_path}/DEBIAN/control

	@echo "Copy mattermost service overrides for omnibus"
	cp mattermost/files/omnibus-overrides.conf ${build_path}/etc/systemd/system/mattermost.service.d/

	@echo "Building Mattermost Omnibus Nightly v${version}-${revision} ${release} package"
ifeq ($(use_docker), true)
	docker run -it -v ${PWD}/build/mattermost-omnibus/${release}:/builder/package -w /builder/package -u ${user_id}:${group_id} debian:latest dpkg-deb --build mattermost-omnibus-nightly_${version}-${revision}
else
	cd ${PWD}/build/mattermost-omnibus/${release}; \
	dpkg-deb --build mattermost-omnibus-nightly_${version}-${revision};
endif
	cp build/mattermost-omnibus/${release}/mattermost-omnibus-nightly_${version}-${revision}.deb mattermost-omnibus-nightly_${version}-${revision}_${release}.deb

	@echo "Build for Mattermost Omnibus Nightly v${version}-${revision} ${release} succeeded"


mattermost-omnibus: bin/mmomni mattermost-omnibus_$(version)-$(revision)_focal.deb mattermost-omnibus_$(version)-$(revision)_jammy.deb mattermost-omnibus_$(version)-$(revision)_noble.deb


mattermost-omnibus-nightly: bin/mmomni mattermost-omnibus-nightly_$(version)-$(revision)_focal.deb mattermost-omnibus-nightly_$(version)-$(revision)_jammy.deb mattermost-omnibus-nightly_$(version)-$(revision)_noble.deb
