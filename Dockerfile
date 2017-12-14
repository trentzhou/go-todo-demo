FROM scratch
ADD go-todo-demo /
ADD static /static
EXPOSE 3000
CMD ["/go-todo-demo"]
