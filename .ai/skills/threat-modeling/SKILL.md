---
name: threat-modeling
description: STRIDE methodology, trust boundaries, abuse cases, attack surface mapping
---
# Threat Modeling

## 1. STRIDE Methodology
Systematically apply STRIDE (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege) to all components. Analyze each interaction against these six threat categories.

## 2. Trust Boundaries
Explicitly diagram and define trust boundaries. A trust boundary is anywhere data moves from a less trusted to a more trusted context. All data crossing a boundary must be strictly validated and sanitized.

## 3. Abuse Cases Definition
Define explicit abuse cases alongside standard use cases. Determine how an attacker might misuse a feature, manipulate inputs, or bypass intended business logic. Design mitigations for each identified abuse case.

## 4. Attack Surface Mapping
Map the entire attack surface. Enumerate all exposed endpoints, background jobs processing external data, file upload processors, and integration points. Minimize this surface area wherever possible.

## 5. Data Flow Diagrams (DFDs)
Create comprehensive Data Flow Diagrams. Trace the lifecycle of sensitive data from ingest to storage to egress. Identify all components that touch sensitive data and ensure they are appropriately hardened.

## 6. Asset Identification
Identify and classify all critical assets. This includes customer data, intellectual property, cryptographic keys, and system availability. Prioritize threat mitigation efforts based on asset criticality.

## 7. Threat Actor Profiling
Define potential threat actors (e.g., script kiddies, malicious insiders, state-sponsored actors). Tailor the threat model and defensive posture to the capabilities and motivations of the expected adversaries.

## 8. Mitigation Verification
Every identified threat must have a documented mitigation strategy. Ensure these mitigations are translated into concrete engineering tasks and verified via automated security testing or manual review.

## 9. Continuous Modeling
Treat threat modeling as an iterative, continuous process. Revisit the model during the design phase of every significant new feature or architectural change. A static threat model is a useless threat model.

## 10. Security Requirements
Derive explicit security requirements from the threat model. Examples include 'All passwords must be hashed using Argon2id' or 'The API must enforce a rate limit of 100 req/min per IP address'.

