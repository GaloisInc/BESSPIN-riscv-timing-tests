#!/usr/bin/env Rscript

library(ggplot2)
library(reshape2)

args = commandArgs(trailingOnly = TRUE)

if (length(args) != 1) {
    stop("need one argument specifying the measurements of instruction timing")
}

data <- read.table(args[1], header = F)
min_val <- min(min(data$V2), min(data$V3), min(data$V4), min(data$V5), min(data$V6), min(data$V7), min(data$V8), min(data$V9))

data$V2 <- data$V2 / min_val
data$V3 <- data$V3 / min_val
data$V4 <- data$V4 / min_val
data$V5 <- data$V5 / min_val
data$V6 <- data$V6 / min_val
data$V7 <- data$V7 / min_val
data$V8 <- data$V8 / min_val
data$V9 <- data$V9 / min_val

colnames(data) <- c("id", "0", "1", "2", "3", "4", "5", "6", "7")
data <- melt(data, id = "id")

ggplot(data, aes(x = variable, y = value)) +
    geom_boxplot(size = 0.2, outlier.size = 0.5, width = 0.2) +
    theme_classic() +
    theme(line = element_blank(),
        panel.grid.major = element_line(colour = "gray70", size = 0.05),
        axis.title = element_text(size = 5),
        axis.text = element_text(size = 5, hjust = 0.5),
        plot.title = element_text(size = 6, hjust = 0.5)) +
	xlab("Number of NO-OP Instructions Modulo 8") +
	ylab("Slowdown")

ggsave("plot.png", width = 6, height = 2)
