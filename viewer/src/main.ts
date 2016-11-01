/**
 * Entry point for the grovewatch browser viewer.
 *
 * It loads a provenance snapshot and renders it. The report source is resolved
 * in this order:
 *   1. a `?report=<url>` query parameter, if present;
