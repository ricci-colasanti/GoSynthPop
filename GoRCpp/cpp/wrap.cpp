#include <Rcpp.h>
using namespace Rcpp;

extern "C" double SumVec(double* vec, int length);

// [[Rcpp::export]]
double sum_go(NumericVector x) {
    return SumVec(REAL(x), x.length());
}