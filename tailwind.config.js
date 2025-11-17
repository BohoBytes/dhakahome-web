module.exports = {
  content: ["./internal/views/**/*.html", "./web/**/*.css"],
  safelist: [
    // Search results page colors
    'bg-[#f2f2f2]',
    'bg-[#f9f9f9]',
    'text-[#414141]',
    'text-[#797979]',
    'text-[#f44335]',
    'border-[#f44335]',
    'bg-[#f44335]',
    // Search results page sizes
    'text-[48px]',
    'text-[24px]',
    'text-[20px]',
    'text-[16px]',
    'max-w-[1723px]',
    'w-[337px]',
    'h-[300px]',
    'min-h-[300px]',
    'rounded-[20px]',
    'rounded-[5px]',
    'gap-[10px]',
    'gap-[20px]',
    'gap-12',
    'px-9',
    'py-[18px]',
    'py-[5px]',
    'h-[37px]',
    'leading-[28px]',
    'leading-[16.8px]',
    'leading-[14.4px]',
    // Shadows
    'shadow-[0px_5px_9.9px_0px_rgba(0,0,0,0.15)]',
    'shadow-[0px_10px_3px_0px_rgba(0,0,0,0),0px_6px_2px_0px_rgba(0,0,0,0.01),0px_4px_2px_0px_rgba(0,0,0,0.05),0px_2px_2px_0px_rgba(0,0,0,0.09),0px_0px_1px_0px_rgba(0,0,0,0.1)]',
  ],
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
