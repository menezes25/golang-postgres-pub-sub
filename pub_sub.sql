DO $$
BEGIN
	-- cria trigger function caso ainda não tenha sido criada
	IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'event_notify') THEN

        CREATE OR REPLACE FUNCTION event_notify() RETURNS trigger AS $FN$
        DECLARE
            payload jsonb;
        BEGIN
			IF TG_OP = 'DELETE' THEN
				payload = jsonb_build_object(
					'op', to_jsonb(TG_OP),
					'payload', to_jsonb(OLD)
				);
			ELSE
				payload = jsonb_build_object(
					'op', to_jsonb(TG_OP),
					'payload', to_jsonb(NEW)
				);
			END IF;

            PERFORM pg_notify('event', payload::TEXT);
            
            RETURN NEW;

        END;
        $FN$ LANGUAGE plpgsql;

	END IF;
	-- cria triggers caso ainda não tenham sido criados 
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_update') THEN
		CREATE TRIGGER event_notify_update AFTER UPDATE ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_insert') THEN
		CREATE TRIGGER event_notify_insert AFTER INSERT ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_delete') THEN
		CREATE TRIGGER event_notify_delete AFTER DELETE ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
END $$;