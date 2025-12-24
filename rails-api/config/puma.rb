# puma config file

# default port that puma will listen on
port ENV.fetch("PORT") { 3000 }

# specifies the env that puma will run in
environment ENV.fetch("RAILS_ENV") { "development" }

# number of threads to use
max_threads_count = ENV.fetch("RAILS_MAX_THREADS") { 5 }
min_threads_count = ENV.fetch("RAILS_MIN_THREADS") { max_threads_count }
threads min_threads_count, max_threads_count

# allow puma to be restarted by bin/rails restart command
plugin :tmp_restart

# specify timeouts
worker_timeout 30

# graceful shutdown config
wait_for_less_busy_worker ENV.fetch("PUMA_WAIT_FOR_LESS_BUSY_WORKER") { 0.001 }.to_f

# graceful shutodwn, wait for requests to finish
on_restart do
  puts "Puma restarting..."
end

lowlevel_error_handler do |e, env, status|
  puts "Puma low-level error: #{e.message}"
  [status, {}, ["Internal Server Error\n"]]
end
