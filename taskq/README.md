# Experiment with taskq and failing tasks

We have a consumer that either
 * succeeds or 
 * fails gracefully (returns an error) or
 * panic()s

You can see that both tasks that fail or tasks that panic() are retries (in tha latter case you need to run the consumer again).
