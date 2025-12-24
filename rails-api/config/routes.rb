Rails.application.routes.draw do
  # Health check endpoint
  get '/health', to: 'health#show'

  get "up" => "rails/health#show", as: :rails_health_check

  # API routes
  namespace :api do
    post 'process', to: 'images#create'
  end
end
