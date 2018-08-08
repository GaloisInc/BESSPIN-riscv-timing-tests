#!/usr/bin/env Rscript

library(geometry)
library(Matrix)

hex_to_dec_dbl <- function(hex_list) {
    ctr <- 1
    transformed_list <- c()

    for (hex_string in hex_list) {
        result <- 0.0
        hex_digits <- strsplit(tolower(hex_string), "")[[1]]

        for (hex_digit in hex_digits) {
            ascii <- utf8ToInt(hex_digit)

            # Check if we have a valid character
            valid <- (ascii >= 48 && ascii <= 57) || (ascii >= 97 && ascii <= 102)

            if (valid == FALSE) {
                stop("trying to convert invalid hex string")
            }

            dec_value <- ascii - 48
            if (ascii >= 97) {
                dec_value <- 10 + ascii - 97
            }

            result = result * 16 + dec_value
        }

        transformed_list[[ctr]] <- result
        ctr <- ctr + 1
    }

    return(transformed_list)
}

dec_dbl_to_hex <- function(dec_dbl, width) {
    result <- ""
    while (width != 0) {
        mod <- dec_dbl %% 16

        mod_s <- intToUtf8(mod + 48)
        if (mod > 9) {
            mod_s <- intToUtf8(mod + 97 - 10)
        }

        dec_dbl <- dec_dbl %/% 16
        result <- paste0(mod_s, result)

        width <- width - 1
    }

    return(result)
}

args = commandArgs(trailingOnly = TRUE)

if (length(args) == 0) {
    stop("need one argument specifying the measurements of instruction timing")
}

data <- read.table(args[1], header = F)
# data$V1 <- hex_to_dec_dbl(data$V1)
# data$V2 <- hex_to_dec_dbl(data$V2)

data <- data.frame(x = data$V1, y = data$V2, z = data$V4)

control <- data.frame(x = data$x, y = data$y)
control_matrix <- as.matrix(control)

point_count <- 100

triangulation <- delaunayn(control_matrix, options = "QbB")
sample_x <- max(data$x) * runif(point_count)
sample_y <- max(data$y) * runif(point_count)
sample_points <- cbind(sample_x, sample_y)

search <- tsearchn(control_matrix, triangulation, sample_points)
b_coords <- search$p
inside_points <- triangulation[search$idx,]

prediction <- c()

for (i in 1:nrow(inside_points)) {
    if (is.na(inside_points[i,1]) == FALSE) {
        sum <- 0

        for (j in 1:ncol(inside_points)) {
            idx <- inside_points[i,j]
            sum <- sum + b_coords[i,j] * data$z[idx]
        }

    } else {
        sum <- NA
    }

    prediction <- c(prediction, sum)
}

result <- cbind(sample_points, prediction)

for (i in 1:nrow(result)) {
    fx <- dec_dbl_to_hex(result[i,1], 16)
    fy <- dec_dbl_to_hex(result[i,2], 16)

    status <- sprintf("%016s %016s %-8.2f\n", fx, fy, result[i,3])
    cat(status)
}
