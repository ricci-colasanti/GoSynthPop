dyn.load("golib.so")

square <- function(x) {
    .C("Square", x = as.integer(x))$x
}

sum_go <- function(arr) {
    if (length(arr) == 0) return(0L)
    arr <- as.integer(arr)  # Force integer type
    .C("Sum", 
       arr = arr, 
       length = as.integer(length(arr)),
       PACKAGE = "golib")$arr
}

# Test cases
print(square(5L))             # Should print 25
print(sum_go(c(1L, 2L, 3L)))  # Should print 6
print(sum_go(integer(0)))     # Should print 0