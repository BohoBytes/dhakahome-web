module.exports = {
  content: ["./internal/views/**/*.html", "./web/**/*.css"],
  theme: {
    extend: {
      colors: {
        primary: "#F44335",
        secondary: "#FFE9E8",
        bg: "#FFFFFF",
        secbg: "#F1F1F1",
        textprimary: "#303030",
        subtext: "#767676",
      },
      fontFamily: {
        sans: [
          "Poppins",
          "ui-sans-serif",
          "system-ui",
          "Segoe UI",
          "Roboto",
          "Helvetica",
          "Arial",
          "Noto Sans",
          "sans-serif",
        ],
      },
      fontSize: {
        // Hero and large text
        hero: ["48px", { lineHeight: "1.1" }], // Reduced from 60px
        section: ["36px", { lineHeight: "1.15" }], // Reduced from 50px
        sectionHeader: ["28px", { lineHeight: "1.2" }], // Reduced from 34px

        // Buttons
        buttonlg: ["20px", { lineHeight: "1.2" }], // Reduced from 26px
        buttonmd: ["18px", { lineHeight: "1.2" }], // Reduced from 22px

        // Content text
        descTitle: ["20px", { lineHeight: "1.35" }], // Reduced from 24px
        subtext: ["16px", { lineHeight: "1.35" }], // Reduced from 22px
        filter: ["14px", { lineHeight: "1.35" }], // Reduced from 16px

        // Navigation and small text
        nav: ["16px", { lineHeight: "1.25" }], // New for navigation
        small: ["12px", { lineHeight: "1.3" }], // New for small text
      },
      boxShadow: { card: "0 4px 20px rgba(0,0,0,0.05)" },
      borderRadius: { xl2: "1rem" },
    },
  },
  plugins: [],
};
