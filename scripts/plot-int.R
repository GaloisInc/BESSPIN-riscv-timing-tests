#!/usr/bin/env Rscript

library(ggplot2)
library(reshape2)

ffs <- function(x) {
    log2(x + 1)
}

args = commandArgs(trailingOnly = TRUE)

if (length(args) != 1) {
    stop("need one argument specifying the file containing timing measurements")
}

path <- args[1]

data <- read.table(path, header = F)

data$V1 <- sapply(data$V1, ffs)
data$V2 <- sapply(data$V2, ffs)
data$V4 <- data$V4 / min(data$V4)

data.mat <- acast(data, V1 ~ V2, value.var = "V4")
data.melt <- melt(data.mat)

zp1 <- ggplot(data.melt, aes(x = Var1, y = Var2, fill = value))
zp1 <- zp1 + labs(title = "latency (ratio)", x = "Most Significant Bit of Operand #1", y = "Most Significant Bit of Operand #2")
zp1 <- zp1 + theme_classic() + theme(line = element_blank(), axis.title = element_text(size = 10), axis.text = element_text(size = 10, hjust = 1), plot.title = element_text(size = 12, hjust = 0.5))
zp1 <- zp1 + geom_tile()
zp1 <- zp1 + scale_fill_gradient2(high = "#F15946", guide = FALSE)
zp1 <- zp1 + scale_x_continuous(breaks = seq(0, 56, by = 8))
zp1 <- zp1 + scale_y_continuous(breaks = seq(0, 56, by = 8))
zp1 <- zp1 + geom_text(aes(Var1, Var2, label = round(value, 1)), size = 3)
zp1 <- zp1 + coord_equal()

ggsave("plot.pdf", width = 4, height = 4, dpi = 420)
