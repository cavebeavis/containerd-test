.PHONY: say_hello generate clean

say_hello:
	@echo "Hello World"

generate:
	@echo "Creating empty text files..."
	touch file-{1..10}.trxt

clean:
	@echo "Cleaning up..."
	rm *.trxt