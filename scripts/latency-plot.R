#!/usr/bin/env Rscript

library(ggplot2)

args = commandArgs(trailingOnly = TRUE)

if (length(args) != 1) {
    stop("need one argument specifying the measurements of instruction timing")
}

data <- read.table(args[1], header = F)
data$V2 <- data$V2 / min(data$V2)

# data <- head(data, 32)
min_x <- min(data$V1)
max_x <- max(data$V1)

ggplot(data, aes(V1, V2)) +
    geom_point(size = 0.05, color = "gray20") +
    scale_x_continuous(breaks = seq(min_x, max_x, by = 64)) +
    theme_classic() +
    theme(line = element_blank(),
        panel.grid.major = element_line(colour = "gray60", size = 0.05),
        axis.title = element_text(size = 5),
        axis.text = element_text(size = 5, hjust = 0.5),
        plot.title = element_text(size = 6, hjust = 0.5)) +
    xlab("Number of NO-OP Instructions") +
    ylab("Slowdown")

ggsave("plot.png", width = 6, height = 2)
