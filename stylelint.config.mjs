/** @type {import("stylelint").Config} */
export default {
  extends: ["stylelint-config-standard", "stylelint-config-tailwindcss"],
  rules: {
    // Tailwind's @apply prelude is a list of utility class names, which
    // stylelint's CSS grammar doesn't recognize as a valid at-rule prelude.
    // stylelint-config-tailwindcss silences at-rule-no-unknown for @apply
    // but not this newer rule (added in stylelint 17.14+), so ignore it here.
    "at-rule-prelude-no-invalid": [
      true,
      { ignoreAtRules: ["apply", "variant", "custom-variant"] },
    ],
  },
};
