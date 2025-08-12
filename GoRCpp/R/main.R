# Load pre-built library
dyn.load("../cpp/sum.so")

# Call the function
result <- .Call("sum_go", c(1.5, 2.5, 3.5))
print(result)  # Should print 7.5