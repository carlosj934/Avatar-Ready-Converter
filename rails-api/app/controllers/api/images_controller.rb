module Api
  class ImagesController < ApplicationController
    # Max file size (10 MB)
    MAX_FILE_SIZE = 10.megabytes

    def create
      # validate file prescence
      unless params[:file].present?
        return render json: { error: 'No file provided'}, status: :bad_request
      end

      uploaded_file = params[:file]

      # validate file sizze
      if uploaded_file.size > MAX_FILE_SIZE
        return render json: { error: 'File too large. Maximum size is 10MB' }, status: :bad_request
      end

      # get processing options
      output_format = params[:output_format] || 'jpg'
      size = params[:size] || 512

      # validate format
      unless ['jpg', 'jpeg', 'png'].include?(output_format.downcase)
        return render json: { error: 'Invalid format. Must be jpg or png' }, status: :bad_request
      end

      begin
        # Forward to Go service
        response = HTTP.timeout(30)
          .post("#{go_service_url}/process",
               form: {
                 file: HTTP::FormData::File.new(uploaded_file.tempfile,
                                                filename: uploaded_file.original_filename,
                                                content_type: uploaded_file.content_type),
                 format: output_format,
                 size: size.to_i
               })

        # check if Go service responded successfully
        unless response.status.success?
          return render json: { error: 'Image processing failed' }, status: :internal_server_error
        end

        # stream the processed image back to the client
        send_data response.body.to_s,
          type: response.content_type.mime_type,
          filename: "avatar.#{output_format}",
          disposition: 'attachment'

        rescue HTTP::TimeoutError
          render json: { error: 'Processing timeout' }, status: :request_timeout
        rescue StandardError => e
          Rails.logger.error "Image processing error: #{e.message}"
          render json: { error: 'An error occurred during processing' }, status: :internal_server_error
        end
    end

    private

    def go_service_url
      ENV.fetch('GO_SERVICE_URL', 'http://localhost:8080')
    end
  end
end

