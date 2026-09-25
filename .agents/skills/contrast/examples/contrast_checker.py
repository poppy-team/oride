#!/usr/bin/env python3
"""Mathematical Color Contrast (WCAG 2.2 & APCA) Calculation Reference."""

import math

def hex_to_rgb(hex_str: str) -> tuple[float, float, float]:
    hex_clean = hex_str.lstrip("#")
    if len(hex_clean) == 3:
        hex_clean = "".join(c * 2 for c in hex_clean)
    r = int(hex_clean[0:2], 16) / 255.0
    g = int(hex_clean[2:4], 16) / 255.0
    b = int(hex_clean[4:6], 16) / 255.0
    return r, g, b

def linearize(c: float) -> float:
    return c / 12.92 if c <= 0.04045 else math.pow((c + 0.055) / 1.055, 2.4)

def relative_luminance(r: float, g: float, b: float) -> float:
    return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b)

def composite_over(fg_hex: str, bg_hex: str, alpha: float) -> str:
    fg_r, fg_g, fg_b = hex_to_rgb(fg_hex)
    bg_r, bg_g, bg_b = hex_to_rgb(bg_hex)
    out_r = fg_r * alpha + bg_r * (1.0 - alpha)
    out_g = fg_g * alpha + bg_g * (1.0 - alpha)
    out_b = fg_b * alpha + bg_b * (1.0 - alpha)
    return f"#{int(out_r * 255):02x}{int(out_g * 255):02x}{int(out_b * 255):02x}"

def wcag_contrast(hex1: str, hex2: str) -> float:
    r1, g1, b1 = hex_to_rgb(hex1)
    r2, g2, b2 = hex_to_rgb(hex2)
    l1 = relative_luminance(r1, g1, b1)
    l2 = relative_luminance(r2, g2, b2)
    lighter = max(l1, l2)
    darker = min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)

def evaluate_wcag_aa(ratio: float, is_large_text: bool = False, is_ui_component: bool = False) -> bool:
    if is_ui_component or is_large_text:
        return ratio >= 3.0
    return ratio >= 4.5

def main() -> None:
    print("=== WCAG 2.2 Color Contrast Invariants Verification ===")

    test_cases = [
        ("Body text on light canvas", "#0f172a", "#ffffff", False, False),
        ("Muted text on light canvas", "#475569", "#ffffff", False, False),
        ("Button text on primary action", "#ffffff", "#2563eb", False, False),
        ("Focus ring on canvas", "#2563eb", "#ffffff", False, True),
        ("Dark theme body text on canvas", "#f8fafc", "#0f172a", False, False),
    ]

    all_passed = True
    for label, fg, bg, is_large, is_ui in test_cases:
        ratio = wcag_contrast(fg, bg)
        passed = evaluate_wcag_aa(ratio, is_large_text=is_large, is_ui_component=is_ui)
        target = "3.0:1" if (is_large or is_ui) else "4.5:1"
        status = "[PASS]" if passed else "[FAIL]"
        print(f"{status} {label:<32}: {ratio:.2f}:1 (Target: >= {target})")
        if not passed:
            all_passed = False

    # Test alpha compositing
    print("\n--- Testing Translucent Alpha Compositing ---")
    semi_transparent_overlay = composite_over("#000000", "#ffffff", 0.7)
    ratio_overlay = wcag_contrast("#ffffff", semi_transparent_overlay)
    print(f"White text over 70% black overlay on white: {ratio_overlay:.2f}:1 (Target: >= 4.5:1)")
    assert ratio_overlay >= 4.5

    assert all_passed, "Some contrast test cases failed!"
    print("\nAll color contrast test cases verified successfully.")

if __name__ == "__main__":
    main()
