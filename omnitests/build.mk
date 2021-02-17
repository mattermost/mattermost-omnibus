omnitests/omnitests.test:
	cd omnitests; \
	./mage build

omnitests-clean:
	cd omnitests; \
	rm -f omnitests.test
