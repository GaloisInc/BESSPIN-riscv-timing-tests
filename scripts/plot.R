library(ggplot2)
library(reshape2)

plot_int_instr_variation <- function(instr) {
    in_file <- paste("../results/data/out.", instr, sep = "")
    data <- read.table(in_file, header = F)

    data$V1 <- data$V1 * 4
    data$V2 <- data$V2 * 4

    data$V3 <- data$V3 / min(data$V3)
    data$V4 <- data$V4 / min(data$V4)

    data.mat <- acast(data, V1 ~ V2, value.var = "V4")
    data.melt <- melt(data.mat)

    zp1 <- ggplot(data.melt, aes(x = Var1, y = Var2, fill = value))
    zp1 <- zp1 + labs(title = paste("latency (ratio) of", instr), x = "Hamming Weight of Operand #1", y = "Hamming Weight of Operand #2")
    zp1 <- zp1 + theme_classic() + theme(line = element_blank(), axis.title = element_text(size = 5), axis.text = element_text(size = 5, hjust = 1), plot.title = element_text(size = 6, hjust = 0.5))
    zp1 <- zp1 + geom_tile()
    zp1 <- zp1 + scale_fill_gradient2(high = "#F15946", guide = FALSE)
    zp1 <- zp1 + scale_x_continuous(breaks = seq(0, 64, by = 8))
    zp1 <- zp1 + scale_y_continuous(breaks = seq(0, 64, by = 8))
    zp1 <- zp1 + geom_text(aes(Var1, Var2, label = round(value, 1)), size = 0.7)
    zp1 <- zp1 + coord_equal()

    ggsave(paste("../results/plots/plot-", instr, ".png", sep = ""), width = 2, height = 2, dpi = 420)
}

plot_fp_instr_variation <- function(instr) {
    in_file <- paste("../results/data/out.", instr, sep = "")
    data <- read.table(in_file, header = F)

    data$V1 <- data$V1 * 1
    data$V2 <- data$V2 * 1

    data$V3 <- data$V3 / min(data$V3)
    data$V4 <- data$V4 / min(data$V4)

    data.mat <- acast(data, V1 ~ V2, value.var = "V4")
    data.melt <- melt(data.mat)

    zp1 <- ggplot(data.melt, aes(x = Var1, y = Var2, fill = value))
    zp1 <- zp1 + labs(title = paste("latency (ratio) of", instr))
    zp1 <- zp1 + theme_classic() + theme(line = element_blank(), axis.title = element_blank(), axis.text.x = element_text(angle = 90, hjust = 1, size = 5), axis.text.y = element_text(size = 5, hjust = 1), plot.title = element_text(size = 6, hjust = 0.5))
    zp1 <- zp1 + geom_tile()
    zp1 <- zp1 + scale_fill_gradient2(high = "#F15946", guide = FALSE)
    zp1 <- zp1 + scale_x_discrete(limits = c(0, 1, 2, 3, 4, 5), labels = c("Zero", "Norm.", "Subn.", "+Inf", "-Inf", "NaN"))
    zp1 <- zp1 + scale_y_discrete(limits = c(0, 1, 2, 3, 4, 5), labels = c("Zero", "Norm.", "Subn.", "+Inf", "-Inf", "NaN"))
    zp1 <- zp1 + geom_text(aes(Var1, Var2, label = round(value, 1)), size = 1.5)
    zp1 <- zp1 + coord_equal()

    ggsave(paste("../results/plots/plot-", instr, ".png", sep = ""), width = 2, height = 2)
}

plot_int_instr_variation("add")
plot_int_instr_variation("and")
plot_int_instr_variation("div")
plot_int_instr_variation("divu")
plot_int_instr_variation("mul")
plot_int_instr_variation("mulh")
plot_int_instr_variation("mulhsu")
plot_int_instr_variation("mulhu")
plot_int_instr_variation("or")
plot_int_instr_variation("rem")
plot_int_instr_variation("remu")
plot_int_instr_variation("sll")
plot_int_instr_variation("slt")
plot_int_instr_variation("sltu")
plot_int_instr_variation("sra")
plot_int_instr_variation("srl")
plot_int_instr_variation("sub")
plot_int_instr_variation("xor")

plot_fp_instr_variation("fmax.s")
plot_fp_instr_variation("fmax.d")
plot_fp_instr_variation("fmin.s")
plot_fp_instr_variation("fmin.d")
plot_fp_instr_variation("fsgnjx.s")
plot_fp_instr_variation("fsgnjx.d")
plot_fp_instr_variation("fsgnjn.s")
plot_fp_instr_variation("fsgnjn.d")
plot_fp_instr_variation("fsgnj.s")
plot_fp_instr_variation("fsgnj.d")
plot_fp_instr_variation("fdiv.s")
plot_fp_instr_variation("fdiv.d")
plot_fp_instr_variation("fmul.s")
plot_fp_instr_variation("fmul.d")
plot_fp_instr_variation("fsub.s")
plot_fp_instr_variation("fsub.d")
plot_fp_instr_variation("fadd.s")
plot_fp_instr_variation("fadd.d")
