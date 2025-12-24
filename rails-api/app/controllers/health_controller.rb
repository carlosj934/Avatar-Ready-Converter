class HealthController < ApplicationController
  def show
    # check if Go service is reachable
    go_healthy = check_go_service

    if go_healthy
      render json: {
        status: 'healthy',
        service: 'rails-api',
        go_service: 'connected',
        timestamp: Time.current.iso8601
      }, status: :ok
    else
      render json: {
        status: 'degraded',
        service: 'rails-api',
        go_service: 'unavailable',
        timestamp: Time.current.iso8601
      }, status: :service_unavailable
    end
  end

  private

  def check_go_service
    response = HTTP.timeout(2).get("#{ENV.fetch('GO_SERVICE_URL', 'http://localhost:8080')}/health")
    response.status.success?
  rescue
    false
  end
end
