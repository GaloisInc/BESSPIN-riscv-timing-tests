#!/usr/bin/env Rscript

library(ggplot2)

args = commandArgs(trailingOnly = TRUE)

if (length(args) != 1) {
    stop("need one argument specifying the measurements of instruction timing")
}

data <- read.table(args[1], header = F)
data$V2 <- data$V2 / min(data$V2)

data <- head(data, 64)
min_x <- min(data$V1)
max_x <- max(data$V1)

ggplot(data, aes(V1, V2)) +
    geom_point(size = 1, color = "red") +
    scale_x_continuous(breaks = seq(min_x, max_x, by = 16)) +
    scale_y_continuous(limits = c(0, 40)) +
    theme_classic() +
    theme(line = element_blank(),
        panel.grid.major = element_line(colour = "gray60", size = 0.05),
        axis.title = element_text(size = 14),
        axis.text = element_text(size = 14, hjust = 0.5),
        plot.title = element_text(size = 16, hjust = 0.5)) +
    xlab("Offset of Start Address of Measurement Code") +
    ylab("Execution Time (Cycles)")

ggsave("plot.pdf", width = 9, height = 3)
