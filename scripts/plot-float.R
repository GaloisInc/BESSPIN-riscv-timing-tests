#!/usr/bin/env Rscript

library(ggplot2)
library(reshape2)

args = commandArgs(trailingOnly = TRUE)

if (length(args) != 1) {
    stop("need one argument specifying the file containing timing measurements")
}

path <- args[1]

data <- read.table(path, header = F)
data$V4 <- data$V4 / min(data$V4)

data.mat <- acast(data, V1 ~ V2, value.var = "V4", fun.aggregate = mean)
data.melt <- melt(data.mat)

zp1 <- ggplot(data.melt, aes(x = Var1, y = Var2, fill = value))
zp1 <- zp1 + labs(title = "latency (ratio)", x = "Operand #1", y = "Operand #2")
zp1 <- zp1 + theme_classic() + theme(line = element_blank(), axis.title = element_text(size = 5), axis.text = element_text(size = 5, hjust = 1), plot.title = element_text(size = 6, hjust = 0.5), axis.text.x = element_text(angle = 90, hjust = 1))
zp1 <- zp1 + geom_tile()
zp1 <- zp1 + scale_fill_gradient2(high = "#F15946", guide = FALSE)
zp1 <- zp1 + geom_text(aes(Var1, Var2, label = round(value, 1)), size = 0.7)
zp1 <- zp1 + coord_equal()

ggsave(paste("plot.png", sep = ""), width = 2, height = 2, dpi = 420)
