class Rack::Attack
  # config cache

  # use rails.cache by default
  Rack::Attack.cache.store = ActiveSupport::Cache::MemoryStore.new

  # Throttle settings
  
  # Throttle all requests by IP (60 requests per minute)
  throttle('req/ip', limit: 60, period: 1.minute) do |req|
    req.ip
  end

  # throttle image processing (10 uploads per minute per IP)
  throttle('uploads/ip', limit: 10, period: 1.minute) do |req|
    req.ip if req.path == '/api/process' && req.post?
  end

  # throttle health check endpoint 
  throttle('health/ip', limit: 30, period: 1.minute) do |req|
    req.ip if req.path == '/health'
  end

  # BlockList and Safelist
  
  # safelist localhost and private IPs
  safelist('allow local') do |req|
    # allow localhost
    '127.0.0.1' == req.ip || '::1' == req.ip ||
    # allow private IP ranges
    req.ip.match(/^10\./) ||                        # 10.0.0.0/8
    req.ip.match(/^172\.(1[6-9]|2\d|3[0-1])\./) ||  # 172.16.0.0/12
    req.ip.match(/^192\.168\./)                     # 192.168.0.0/16
  end

  # Custom response for throttle requests

  self.throttled_responder = lambda do |env|
    match_data = env['rack.attack.match_data']
    now = match_data[:epoch_time]

    # calculate when the limit will reset
    retry_after = match_data[:period] - (now % match_data[:period])

    [
      429, # HTTP 429 Too Many Requests
      {
        'Content-Type' => 'application/json',
        'Retry-After' => retry_after.to_s,
        'X-RateLimit-Limit' => match_data[:limit].to_s,
        'X-RateLimit-Remaining' => '0',
        'X-RateLimit-Reset' => (now + retry_after).to_s
      },
      [{ 
        error: 'Rate limit exceeded',
        message: 'Too many requests. Please try again later.',
        retry_after: retry_after
      }.to_json]
    ]
  end

  # Logging

  # log blocked/throttled requests
  ActiveSupport::Notifications.subscribe(/rack_attack/) do |name, start, finish, request_id, payload|
    req = payload[:request]

    if name == 'rack_attack.throttle'
      Rails.logger.warn("[Rack::Attack][Throttled] IP: #{req.ip} | Path: #{req.path} | Matched: #{req.env['rack.attack.matched']}")
    elsif name == 'rack_attack.blocklist'
      Rails.logger.warn("[Rack::Attack][Blocked] IP: #{req.ip} | Path: #{req.path}")
    end
  end
end
